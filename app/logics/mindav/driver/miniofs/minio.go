package miniofs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/totoval/framework/config"
	"github.com/totoval/framework/helpers/log"
	"github.com/totoval/framework/helpers/toto"
	"github.com/totoval/framework/helpers/zone"
	"golang.org/x/net/webdav"
	"os"
	"path"
	"strings"
)

const KEEP_FILE_NAME = ".mindavkeep"
const KEEP_FILE_CONTENT_TYPE = "application/mindav-folder-keeper"

type minioFileSystem struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	client          *minio.Client
	bucketName      string
	location        string
	rootInfo        *fileInfo
}

func New(bucketName string, location string) *minioFileSystem {
	m := &minioFileSystem{
		Endpoint:        config.GetString("webdav.filesystems.minio.endpoint"),
		AccessKeyID:     config.GetString("webdav.filesystems.minio.access_key_id"),
		SecretAccessKey: config.GetString("webdav.filesystems.minio.secret_access_key"),
		UseSSL:          config.GetBool("webdav.filesystems.minio.use_ssl"),
		bucketName:      bucketName,
		location:        location,
		rootInfo: &fileInfo{minio.ObjectInfo{
			Key:          "/",
			Size:         0,
			LastModified: zone.Now(),
			ContentType:  "inode/directory",
			ETag:         "",
			StorageClass: "",
		}},
	}

	var err error
	if m.client, err = minio.New(m.Endpoint, m.AccessKeyID, m.SecretAccessKey, m.UseSSL); err != nil {
		panic(err)
	}

	err = m.MkBucket()
	if err != nil {
		panic(err)
	}

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
func (m *minioFileSystem) MkBucket() (err error) {
	exists, err := m.client.BucketExists(m.bucketName)
	if err != nil {
		return log.Error(err, toto.V{"op": "mkbucket check"})
	}

	if exists {
		log.Info("We already own", toto.V{"bucket": m.bucketName})
		return nil
	}

	// not exist
	if err := m.client.MakeBucket(m.bucketName, m.location); err != nil {
		return log.Error(err, toto.V{"op": "mkbucket make"})
	}

	log.Info("Successfully created", toto.V{"bucket": m.bucketName})
	return nil
}
func (m *minioFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {

	name, err := clearName(name)
	if err != nil {
		return err
	}

	fileBytes := bytes.NewBuffer([]byte{})
	_, err = m.client.PutObject(m.bucketName, strings.TrimPrefix(path.Join(name, KEEP_FILE_NAME), "/"), bytes.NewBuffer([]byte{}), int64(fileBytes.Len()), minio.PutObjectOptions{ContentType: KEEP_FILE_CONTENT_TYPE})
	if err != nil {
		return log.Error(err, toto.V{"op": "mkdir", "name": path.Join(name, KEEP_FILE_NAME)})
	}
	log.Info("mkdir success", toto.V{"name": name})
	return nil
}
func (m *minioFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {

	name, err := clearName(name)
	if err != nil {
		return nil, err
	}

	log.Trace("minio openfile", toto.V{"Name": name})

	if len(name) == 0 {
		return file{m, nil, "/"}, nil
	}

	// file
	object, err := m.client.GetObject(m.bucketName, strings.TrimPrefix(name, "/"), minio.GetObjectOptions{})
	log.Trace("open file", toto.V{"name": name})
	if err != nil {
		return nil, err
	}

	return file{m, object, name}, nil
}
func (m *minioFileSystem) RemoveAll(ctx context.Context, name string) error {

	name, err := clearName(name)
	if err != nil {
		return err
	}

	log.Trace("minio removeall", toto.V{"Name": name})

	objectsCh := make(chan string)
	// Send object names that are needed to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range m.client.ListObjects(m.bucketName, name, true, nil) {
			if object.Err != nil {
				_ = log.Error(object.Err, toto.V{"op": "removeAll", "name": name})
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
func (m *minioFileSystem) Rename(ctx context.Context, oldName, newName string) error {

	oldParentName, err := clearName(oldName)
	if err != nil {
		return err
	}
	newParentName, err := clearName(newName)
	if err != nil {
		return err
	}

	log.Trace("minio rename", toto.V{"Old": oldName, "New": newName, "oldParentName": oldParentName, "newParentName": newParentName})

	//newName = strings.Replace(newName, path.Dir(oldName), "", 1)
	err = m.walkDir(ctx, oldParentName, newParentName, oldName)
	if err != nil {
		return err
	}

	// return nil // for test
	return m.RemoveAll(ctx, oldName)
}
func (m *minioFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {

	name, err := clearName(name)
	if err != nil {
		return nil, err
	}

	log.Trace("minio stat", toto.V{"Name": name})
	if len(name) == 0 {
		// root dir
		return m.rootInfo, nil
	}

	stat, err := m.client.StatObject(m.bucketName, name, minio.StatObjectOptions{})
	if err != nil {
		if _err, ok := err.(minio.ErrorResponse); ok {
			if _err.Code == "NoSuchKey" {
				// check is dir
				if !m.isDir(name) {
					// not exist
					return nil, os.ErrNotExist
				}

				// is dir
				theName, err := clearName(name)
				if err != nil {
					return nil, err
				}
				return fileInfo{minio.ObjectInfo{
					Key:          theName,
					Size:         0,
					LastModified: zone.Now(),
					ContentType:  "inode/directory",
					ETag:         "",
					StorageClass: "",
				}}, nil
			}
		}
		return nil, log.Error(err)
	}
	return fileInfo{stat}, nil
}
func (m *minioFileSystem) walkDir(ctx context.Context, oldParentName, newParentName, oldName string) error {

	oldNameTrim := strings.Trim(oldName, "/")
	newName := newParentName
	if strings.Contains(oldNameTrim, "/") {
		// has child dirs
		newName = strings.Replace(oldName, oldParentName, newParentName, 1)
	}

	log.Debug("walkDir", toto.V{"oldParentName": oldParentName, "newParentName": newParentName, "oldName": oldName, "newName": newName, "isDir": m.isDir(oldName)})

	if !m.isDir(oldName) {
		src := minio.NewSourceInfo(m.bucketName, strings.TrimPrefix(oldName, "/"), nil)
		dst, err := minio.NewDestinationInfo(m.bucketName, strings.TrimPrefix(newName, "/"), nil, nil)
		if err != nil {
			return log.Error(err, toto.V{"op": "walkDir", "dst": dst})
		}
		if err := m.client.CopyObject(dst, src); err != nil {
			return log.Error(err, toto.V{"op": "walkDir", "old": oldName, "new": newName})
		}

		return nil
	}

	// is dir, then readdir
	minioObj, err := m.OpenFile(ctx, oldName, 0, 777)
	if err != nil {
		return log.Error(err, toto.V{"op": "OpenFile", "old": oldName, "new": newName})
	}
	oldFileDirChildren, err := minioObj.Readdir(-1)
	if err != nil {
		return err
	}
	for _, child := range oldFileDirChildren {
		log.Debug("walkDir oldFileDirChildren", toto.V{"op": "walkDir", "oldName": oldName, "child": child.Name(), "len": len(oldFileDirChildren)})
		if err := m.walkDir(ctx, oldName, newName, path.Join(oldName, child.Name())); err != nil {
			return err
		}
	}
	return nil
}
func (m *minioFileSystem) isDir(name string) bool {
	if !strings.HasSuffix(name, "/") {
		name = name + "/"
	}

	//@todo cache result
	//cache.Put()

	childrenCount := 0
	for obj := range m.client.ListObjectsV2(m.bucketName, name, false, nil) {
		if obj.Err != nil {
			_ = log.Error(obj.Err)
			return false
		}
		childrenCount++
	}

	log.Trace("isDir", toto.V{"name": name, "childrenCount": childrenCount})

	if childrenCount <= 0 {
		// not dir, not exist

		//double check dir, if it contains hidden .mindavkeep file
		_, err := m.client.StatObject(m.bucketName, path.Join(name, KEEP_FILE_NAME), minio.StatObjectOptions{})
		if err != nil {
			// not dir or not exist
			return false
		}

		// empty dir
		return true
	} else {
		// not empty dir
		return true
	}
}
