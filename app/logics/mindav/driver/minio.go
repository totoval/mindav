package driver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/totoval/framework/config"
	log2 "github.com/totoval/framework/helpers/log"
	"github.com/totoval/framework/helpers/toto"
	"github.com/totoval/framework/helpers/zone"
	"golang.org/x/net/webdav"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"
)
const KEEP_FILE_NAME = ".mindavkeep"
//const KEEP_FILE_CONTENT_TYPE = "application/mindav-folder-keeper"
const KEEP_FILE_CONTENT_TYPE = "application/octet-stream"

type Minio struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	client          *minio.Client
	bucketName      string
	location        string
}

func NewMinio(bucketName string, location string) Minio {
	m := Minio{
		Endpoint:        config.GetString("webdav.filesystems.minio.endpoint"),
		AccessKeyID:     config.GetString("webdav.filesystems.minio.access_key_id"),
		SecretAccessKey: config.GetString("webdav.filesystems.minio.secret_access_key"),
		UseSSL:          config.GetBool("webdav.filesystems.minio.use_ssl"),
		bucketName:      bucketName,
		location:        location,
	}

	var err error
	if m.client, err = minio.New(m.Endpoint, m.AccessKeyID, m.SecretAccessKey, m.UseSSL); err != nil {
		panic(err)
	}

	err = m.MkBucket()
	if err != nil {
		panic(err)
	}
	// if exists {
	// 	log.Println("bucket "+bucketName+"@"+location+" has already exists, skip creating")
	// }

	return m
}
func clearName(name string) (string, error) {
	if name == "/" {
		return "", nil
	}
	slashed := strings.HasSuffix(name, "/")
	name = path.Clean(name)
	if !strings.HasSuffix(name, "/") && slashed {
		name += "/"
	}
	if !strings.HasPrefix(name, "/") {
		return "", os.ErrInvalid
	}
	return name, nil
}
func (m Minio) MkBucket() (err error) {
	exists, err := m.client.BucketExists(m.bucketName)
	if err != nil {
		log.Println(err)
		return err
	}

	if exists {
		log.Printf("We already own %s\n", m.bucketName)
		return nil
	}

	// not exist
	if err := m.client.MakeBucket(m.bucketName, m.location); err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Successfully created %s\n", m.bucketName)
	return nil
}
func (m Minio) Mkdir(ctx context.Context, name string, perm os.FileMode) error {

	name, err := clearName(name)
	if err != nil {
		return err
	}

	fileBytes := bytes.NewBuffer([]byte{})
	n, err := m.client.PutObject(m.bucketName, strings.TrimPrefix(path.Join(name, KEEP_FILE_NAME), "/"), bytes.NewBuffer([]byte{}), int64(fileBytes.Len()), minio.PutObjectOptions{ContentType: KEEP_FILE_CONTENT_TYPE})
	if err != nil {
		log2.Error(err, toto.V{"op": "mkdir", "name":path.Join(name, KEEP_FILE_NAME)})
		return err
	}
	log.Println("Uploaded", KEEP_FILE_NAME, " of size: ", n, "Successfully.")
	return nil
}
func (m Minio) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {

	name, err := clearName(name)
	if err != nil {
		return nil, err
	}

	log2.Trace("minio openfile", toto.V{"Name": name})

	if len(name) == 0 {
		return file{&m, nil, "/"}, nil
	}

	// file
	object, err := m.client.GetObject(m.bucketName, strings.TrimPrefix(name, "/"), minio.GetObjectOptions{})
	log.Println("OPEN---------------------------------" + name)
	if err != nil {
		return nil, err
	}

	return file{&m, object, name}, nil
}
func (m Minio) RemoveAll(ctx context.Context, name string) error {

	name, err := clearName(name)
	if err != nil {
		return err
	}

	log2.Trace("minio removeall", toto.V{"Name":name})

	objectsCh := make(chan string)
	// Send object names that are needed to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range m.client.ListObjects(m.bucketName, name, true, nil) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			objectsCh <- object.Key
		}
	}()

	for rErr := range m.client.RemoveObjectsWithContext(ctx, m.bucketName, objectsCh) {
		fmt.Println("Error detected during deletion: ", rErr)

		if rErr.Err != nil {
			return rErr.Err
		}
	}

	return m.client.RemoveObject(m.bucketName, name)
}
func (m Minio) Rename(ctx context.Context, oldName, newName string) error {

	oldParentName, err := clearName(oldName)
	if err != nil {
		return err
	}
	newParentName, err := clearName(newName)
	if err != nil {
		return err
	}

	log2.Trace("minio rename", toto.V{"Old":oldName, "New":newName, "oldParentName":oldParentName, "newParentName":newParentName})

	//newName = strings.Replace(newName, path.Dir(oldName), "", 1)
	err = m.walkDir(ctx, oldParentName, newParentName, oldName)
	if err != nil {
		return err
	}

	// return nil // for test
	return m.RemoveAll(ctx, oldName)
}
func (m Minio) Stat(ctx context.Context, name string) (os.FileInfo, error) {

	name, err := clearName(name)
	if err != nil {
		return nil, err
	}

	log2.Trace("minio stat", toto.V{"Name":name})
	if len(name) == 0 {
		//@todo root dir
		return fileInfo{minio.ObjectInfo{
			Key:          "/",
			Size:         0,
			LastModified: zone.Now(),
			ContentType:  "inode/directory",
			ETag:  "",
			StorageClass: "",
		}}, nil
	}

	stat, err := m.client.StatObject(m.bucketName, name, minio.StatObjectOptions{})
	if err != nil {
		if _err, ok := err.(minio.ErrorResponse); ok {
			if _err.Code == "NoSuchKey" {
				// check is dir
				if !m.isDir(name){
					// not exist
					return nil, os.ErrNotExist
				}

				// is dir
				theName, err := clearName(name)
				if err != nil{
					return nil, err
				}
				return fileInfo{minio.ObjectInfo{
					Key:          theName,
					Size:         0,
					LastModified: zone.Now(),
					ContentType:  "inode/directory",
					ETag:  "",
					StorageClass: "",
				}}, nil
			}
		}
		return nil, log2.Error(err)
	}
	return fileInfo{stat}, nil
}
func (m Minio) walkDir(ctx context.Context, oldParentName, newParentName, oldName string) error {

	oldNameTrim := strings.Trim(oldName, "/")
	newName := newParentName
	if strings.Contains(oldNameTrim, "/"){
		// has child dirs
		newName = strings.Replace(oldName, oldParentName, newParentName, 1)
	}

	log2.Debug("walkDir", toto.V{"oldParentName":oldParentName, "newParentName":newParentName, "oldName": oldName, "newName": newName, "isDir":m.isDir(oldName)})

	if !m.isDir(oldName) {
		src := minio.NewSourceInfo(m.bucketName, strings.TrimPrefix(oldName, "/"), nil)
		dst, err := minio.NewDestinationInfo(m.bucketName, strings.TrimPrefix(newName, "/"), nil, nil)
		if err != nil {
			return log2.Error(err, toto.V{"op":"walkDir", "dst":dst})
		}
		if err := m.client.CopyObject(dst, src); err != nil {
			return log2.Error(err, toto.V{"op":"walkDir", "old":oldName, "new":newName})
		}

		return nil
	}


	// is dir, then readdir
	minioObj, err := m.OpenFile(ctx, oldName, 0, 777)
	if err != nil {
		return log2.Error(err, toto.V{"op":"OpenFile", "old":oldName, "new":newName})
	}
	oldFileDirChildren, err := minioObj.Readdir(-1)
	if err != nil {
		return err
	}
	for _, child := range oldFileDirChildren {
		log2.Debug("walkDir oldFileDirChildren", toto.V{"op": "walkDir","oldName":oldName, "child":child.Name(), "len":len(oldFileDirChildren)})
		if err := m.walkDir(ctx, oldName, newName, path.Join(oldName, child.Name())); err != nil {
			return err
		}
	}
	return nil
}
func (m Minio) isDir(name string) bool {
	if !strings.HasSuffix(name, "/"){
		name = name + "/"
	}

	//@todo cache result
	//cache.Put()

	childrenCount := 0
	for obj := range m.client.ListObjectsV2(m.bucketName, name, false, nil){
		if obj.Err != nil{
			_ = log2.Error(obj.Err)
			return false
		}
		childrenCount++
	}

	log2.Trace("isDir", toto.V{"name":name, "childrenCount":childrenCount})

	if childrenCount <= 0{
		// not dir, not exist

		//double check dir, if it contains hidden .mindavkeep file
		_, err := m.client.StatObject(m.bucketName, path.Join(name, KEEP_FILE_NAME), minio.StatObjectOptions{})
		if err != nil{
			// not dir or not exist
			return false
		}

		// empty dir
		return true
	}else{
		// not empty dir
		return true
	}
}

type file struct {
	m *Minio
	*minio.Object
	name string
}


func (mo file) Stat() (os.FileInfo, error) {
	log2.Trace("file stat", toto.V{"name":mo.name})
	return mo.m.Stat(context.Background(), mo.name)
}
func (mo file) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = mo.m.client.PutObject(mo.m.bucketName, strings.TrimPrefix(mo.name, "/"), r, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {

		return 0, log2.Error(err, toto.V{"op":"ReadFrom", "name":mo.name})
	}
	fmt.Println("Successfully uploaded bytes: ", n)
	return n, nil
}
func (mo file) Write(p []byte) (n int, err error) {
	return len(p), nil // useless
}

func (mo file) Readdir(count int) (fileInfoList []os.FileInfo, err error) {
	log.Println(" --- file Readdir " + mo.name)

	name, err := clearName(mo.name)
	if err != nil {
		return nil, err
	}

	if name != "" {
		if !strings.HasSuffix(name, "/") {
			name = name + "/"
		}
	}

	// Create a done channel to control 'ListObjects' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)
	// List all objects from a bucket-name with a matching prefix.
	for object := range mo.m.client.ListObjectsV2(mo.m.bucketName, name, false, doneCh) {
		err = object.Err
		if err != nil {
			fmt.Println(object.Err)
			// return
			break
		}

		if object.StorageClass == "" && object.ETag == "" && object.Size == 0 {
			object.ContentType = "inode/directory"
		}

		fileInfoList = append(fileInfoList, fileInfo{object})
	}

	return fileInfoList, err
}

// func (m Minio)ReadDir(name string){
// 	// Create a done channel to control 'ListObjectsV2' go routine.
// 	doneCh := make(chan struct{})
//
// 	// Indicate to our routine to exit cleanly upon return.
// 	defer close(doneCh)
//
// 	isRecursive := false
// 	objectCh := m.client.ListObjectsV2(m.bucketName, name, isRecursive, doneCh)
// 	for object := range objectCh {
// 		if object.Err != nil {
// 			fmt.Println(object.Err)
// 			return
// 		}
// 		fmt.Println(object)
// 	}
// }





















type fileInfo struct {
	minio.ObjectInfo
}

func (moi fileInfo) Name() string {
	name := moi.ObjectInfo.Key

	name = strings.Trim(name, "/")

	if strings.Contains(name, "/") {
		name = path.Clean(strings.Replace(name, path.Dir(name), "", 1))
	}
	log2.Info(name, toto.V{"Key":moi.ObjectInfo.Key, "ObjectName": name})


	return name
} // base name of the file
func (moi fileInfo) Size() int64 {
	return moi.ObjectInfo.Size
} // length in bytes for regular files; system-dependent for others
func (moi fileInfo) Mode() os.FileMode {
	return 777
} // file mode bits
func (moi fileInfo) ModTime() time.Time {
	return moi.ObjectInfo.LastModified
} // modification time
func (moi fileInfo) IsDir() bool {
	// object.StorageClass == "" && object.ETag == "" && object.Size == 0
	isDir := /*strings.HasSuffix(moi.ObjectInfo.Key, "/") || */moi.ObjectInfo.ContentType == "inode/directory"
	log2.Warn(moi.ObjectInfo.Key, toto.V{"IsDir": isDir})
	// return moi.ObjectInfo.ContentType == "inode/directory"
	return isDir
}// abbreviation for Mode().IsDir()
func (moi fileInfo) Sys() interface{} {
	return nil
} // underlying data source (can return nil)
