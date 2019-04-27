package driver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go"
	"github.com/totoval/framework/config"
	"golang.org/x/net/webdav"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Minio struct {
	Endpoint string
	AccessKeyID string
	SecretAccessKey string
	UseSSL bool
	client *minio.Client
	bucketName string
	location string
}

func NewMinio(bucketName string, location string) Minio {
	m := Minio{
		Endpoint: config.GetString("webdav.filesystems.minio.endpoint"),
		AccessKeyID: config.GetString("webdav.filesystems.minio.access_key_id"),
		SecretAccessKey: config.GetString("webdav.filesystems.minio.secret_access_key"),
		UseSSL: config.GetBool("webdav.filesystems.minio.use_ssl"),
		bucketName: bucketName,
		location: location,
	}

	var err error
	if m.client, err = minio.New(m.Endpoint, m.AccessKeyID, m.SecretAccessKey, m.UseSSL); err != nil{
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
func (m Minio) resolve() string {
	return ""
}
func (m Minio)Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	const KEEP_FILE_NAME = ".mindavkeep"

	if name != "/"{
		name = strings.TrimPrefix(name, "/")
	}

	fileBytes := bytes.NewBuffer([]byte{})
	n, err := m.client.PutObject(m.bucketName, name + KEEP_FILE_NAME, bytes.NewBuffer([]byte{}), int64(fileBytes.Len()), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Uploaded", KEEP_FILE_NAME, " of size: ", n, "Successfully.")
	return nil
}
func (m Minio)OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error){
	log.Println(" --- Minio OpenFile "+name)

	if name != "/"{
		name = strings.TrimPrefix(name, "/")
	}

	if name == "/" {
		return minioObject{&m, nil, name}, nil
	}


	// file
	object, err := m.client.GetObject(m.bucketName, name, minio.GetObjectOptions{})
	log.Println("OPEN---------------------------------"+name)
	if err != nil {
		return nil, err
	}

	return minioObject{&m, object, name}, nil
}
func(m Minio)RemoveAll(ctx context.Context, name  string) error {
	return nil
}
func (m Minio)Rename(ctx context.Context, oldName, newName string) error {
	return nil
}

type minioObject struct {
	m *Minio
	*minio.Object
	name string
}

func (mo minioObject) Stat() (os.FileInfo, error) {
	log.Println(" --- minioObject Stat "+mo.name)

	_, err := mo.m.client.StatObject(mo.m.bucketName, mo.name, minio.StatObjectOptions{})
	if err != nil {
		log.Println("minioObject stat", err)
		if _err, ok := err.(minio.ErrorResponse); ok {
			if _err.Code == "NoSuchKey" { // @todo maybe file 404 either
				// dir

				// fmt.Println("---------------------------------------")
				// if !strings.HasSuffix(mo.name, "/"){
				// 	mo.name = mo.name + "/"
				// }
				// fl, err := mo.Readdir(0)
				// debug.Dump(fl, err)

				return minioObjectInfo{minio.ObjectInfo{
					Key: mo.name,
					Size: 0,
					LastModified: time.Now(),
					ContentType: "inode/directory",
				}}, nil
			}
		}
		return nil, err
	}
	// if strings.HasSuffix(mo.name, "/"){
	// 	return minioObjectInfo{minio.ObjectInfo{
	// 			Key: mo.name,
	// 			Size: 0,
	// 			LastModified: time.Now(),
	// 			ContentType: "inode/directory",
	// 	}}, nil
	// }
	oi, err := mo.Object.Stat()
	return minioObjectInfo{oi}, err
}
func (mo minioObject)ReadFrom(r io.Reader) (n int64, err error) {
	n, err = mo.m.client.PutObject(mo.m.bucketName, mo.name, r, -1, minio.PutObjectOptions{ContentType:"application/octet-stream"})
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	fmt.Println("Successfully uploaded bytes: ", n)
	return n, nil
}
func (mo minioObject) Write(p []byte) (n int, err error) {
	return len(p), nil //useless
}

func(mo minioObject)Readdir(count int)(fileInfoList []os.FileInfo, err error){
	log.Println(" --- minioObject Readdir "+mo.name)
	name := strings.TrimPrefix(mo.name, "/")
	// Create a done channel to control 'ListObjects' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)
	// List all objects from a bucket-name with a matching prefix.
	for object := range mo.m.client.ListObjectsV2(mo.m.bucketName, name, false, doneCh) {
		err = object.Err
		if err != nil {
			fmt.Println(object.Err)
			//return
			break
		}
		fileInfoList = append(fileInfoList, minioObjectInfo{object})
	}
	return fileInfoList, err
}
func (m Minio)Stat(ctx context.Context, name string) (os.FileInfo, error) {
	log.Println(" --- Minio Stat "+name)

	if name != "/"{
		name = strings.TrimPrefix(name, "/")
	}

	stat, err := m.client.StatObject(m.bucketName, name, minio.StatObjectOptions{})
	if err != nil {
		log.Println(err)
		if _err, ok := err.(minio.ErrorResponse); ok {
			if _err.Code == "NoSuchKey" {
				// dir
				return minioObjectInfo{minio.ObjectInfo{
					Key: name,
					Size: 0,
					LastModified: time.Now(),
					ContentType: "inode/directory",
				}}, nil
			}
		}
		return nil, err
	}
	log.Println(stat)
	return minioObjectInfo{stat}, nil
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

type minioObjectInfo struct {
	minio.ObjectInfo
}
func (moi minioObjectInfo)Name() string {
	n := strings.TrimSuffix(moi.ObjectInfo.Key, "/")
	t := strings.Split(n, "/")
	return t[len(t)-1]//moi.ObjectInfo.Key
}       // base name of the file
func (moi minioObjectInfo)Size() int64 {
	return moi.ObjectInfo.Size
}       // length in bytes for regular files; system-dependent for others
func (moi minioObjectInfo)Mode() os.FileMode {
	return 777
}    // file mode bits
func (moi minioObjectInfo)ModTime() time.Time {
	return moi.ObjectInfo.LastModified
}// modification time
func (moi minioObjectInfo)IsDir() bool {
	fmt.Println(moi.Name()+" isdir ", moi.ObjectInfo.ContentType == "inode/directory")
	return moi.ObjectInfo.ContentType == "inode/directory"
}       // abbreviation for Mode().IsDir()
func (moi minioObjectInfo)Sys() interface{} {
	return nil
}  // underlying data source (can return nil)