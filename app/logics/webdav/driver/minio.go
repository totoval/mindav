package driver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go"
	"github.com/totoval/framework/config"
	"github.com/totoval/framework/helpers/debug"
	"golang.org/x/net/webdav"
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

func (m Minio)Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	const KEEP_FILE_NAME = ".keep"

	name = strings.TrimPrefix(name, "/")

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
	if strings.HasSuffix(name, "/") {
		return MinioObject{&m, nil, name}, nil
	}

	// file
	object, err := m.client.GetObject(m.bucketName, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return MinioObject{&m, object, name}, nil
}
func(m Minio)RemoveAll(ctx context.Context, name  string) error {
	return nil
}
func (m Minio)Rename(ctx context.Context, oldName, newName string) error {
	return nil
}

type MinioObject struct {
	m *Minio
	*minio.Object
	name string
}

func (mo MinioObject) Stat() (os.FileInfo, error) {

	if strings.HasSuffix(mo.name, "/"){
		return objectInfo{minio.ObjectInfo{
				Key: mo.name,
				Size: 0,
				LastModified: time.Now(),
				ContentType: "inode/directory",
		}}, nil
	}
	oi, err := mo.Object.Stat()
	return objectInfo{oi}, err
}

func (mo MinioObject) Write(p []byte) (n int, err error) {
	panic("implement me")
	return 0, nil //@todo
}

func(mo MinioObject)Readdir(count int)(fileInfoList []os.FileInfo, err error){
	// Create a done channel to control 'ListObjects' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)
	// List all objects from a bucket-name with a matching prefix.
	for object := range mo.m.client.ListObjects(mo.m.bucketName, mo.name, false, doneCh) {
		err = object.Err
		if err != nil {
			fmt.Println(object.Err)
			//return
			break
		}
		fileInfoList = append(fileInfoList, objectInfo{object})
	}
	return fileInfoList, err
}
func (m Minio)Stat(ctx context.Context, name string) (os.FileInfo, error) {
	if strings.HasSuffix(name, "/") {
		// dir
		return objectInfo{minio.ObjectInfo{
			Key: name,
			Size: 0,
			LastModified: time.Now(),
			ContentType: "inode/directory",
		}}, nil
	}

	stat, err := m.client.StatObject(m.bucketName, name, minio.StatObjectOptions{})
	debug.Dump(ctx, name, stat, err)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(stat)
	return objectInfo{stat}, nil
}
func (m Minio)ReadDir(name string){
	// Create a done channel to control 'ListObjectsV2' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	isRecursive := false
	objectCh := m.client.ListObjectsV2(m.bucketName, name, isRecursive, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}
		fmt.Println(object)
	}
}

type objectInfo struct {
	minio.ObjectInfo
}
func (oi objectInfo)Name() string {
	return oi.ObjectInfo.Key
}       // base name of the file
func (oi objectInfo)Size() int64 {
	return oi.ObjectInfo.Size
}       // length in bytes for regular files; system-dependent for others
func (oi objectInfo)Mode() os.FileMode {
	return 777
}    // file mode bits
func (oi objectInfo)ModTime() time.Time {
	return oi.ObjectInfo.LastModified
}// modification time
func (oi objectInfo)IsDir() bool {
	fmt.Println(oi.ObjectInfo.ContentType)
	return oi.ObjectInfo.ContentType == "inode/directory"
}       // abbreviation for Mode().IsDir()
func (oi objectInfo)Sys() interface{} {
	return nil
}  // underlying data source (can return nil)