package minio

import (
	minio_package "github.com/minio/minio-go"
	"github.com/totoval/framework/config"
	"github.com/totoval/framework/helpers/log"
	"github.com/totoval/framework/helpers/toto"
	"golang.org/x/net/webdav"
	"context"
	"os"
	"path"
	"strings"
)

type minio struct {
	Endpoint string
	AccessKeyID string
	SecretAccessKey string
	UseSSL bool
	client *minio_package.Client
	bucketName string
	location string
}
func New(bucketName string, location string) *minio {
	m := &minio{
		Endpoint: config.GetString("webdav.filesystems.minio.endpoint"),
		AccessKeyID: config.GetString("webdav.filesystems.minio.access_key_id"),
		SecretAccessKey: config.GetString("webdav.filesystems.minio.secret_access_key"),
		UseSSL: config.GetBool("webdav.filesystems.minio.use_ssl"),
		bucketName: bucketName,
		location: location,
	}

	var err error
	if m.client, err = minio_package.New(m.Endpoint, m.AccessKeyID, m.SecretAccessKey, m.UseSSL); err != nil{
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

func (m *minio) MkBucket() (err error) {
	exists, err := m.client.BucketExists(m.bucketName)
	if err != nil {
		return log.Error(err)
	}

	if exists {
		log.Info("We already own", toto.V{"bucketName": m.bucketName})
		return nil
	}

	// not exist
	if err := m.client.MakeBucket(m.bucketName, m.location); err != nil {
		return log.Error(err)
	}

	log.Info("Successfully created", toto.V{"bucketName":m.bucketName})
	return nil
}

func (m *minio)Mkdir(ctx context.Context, fullDirPath string, perm os.FileMode) error {
	log.Info("Mkdir", toto.V{"fullDirPath": fullDirPath, "perm":perm})

	panic(1)
}
func (m *minio)OpenFile(ctx context.Context, fullFilePath string, flag int, perm os.FileMode) (webdav.File, error){
	log.Info("OpenFile", toto.V{"fullFilePath": fullFilePath, "flag": flag, "perm":perm})

	var err error
	if fullFilePath, err = clearName(fullFilePath); err != nil {
		return nil, err
	}

	panic(fullFilePath)
}
func(m *minio)RemoveAll(ctx context.Context, fullFilePath string) error {
	log.Info("RemoveAll", toto.V{"fullFilePath": fullFilePath})

	panic(3)
}
func (m *minio)Rename(ctx context.Context, oldName, newName string) error {
	log.Info("Rename", toto.V{"oldName": oldName, "newName": newName})

	panic(4)
}
func (m *minio)Stat(ctx context.Context, fullFilePath string) (os.FileInfo, error) {
	log.Info("Stat", toto.V{"fullFilePath": fullFilePath})

	panic(5)
}