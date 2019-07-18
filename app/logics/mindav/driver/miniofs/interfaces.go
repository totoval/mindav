package miniofs

import (
	"context"
	"github.com/minio/minio-go/v6"
	"github.com/totoval/framework/helpers/zone"
	"io"
	"net/url"
)

type MinioClienter interface {
	MinioLser
	MinioUploader
	MinioStater
	MinioGetter
}

type MinioLser interface {
	ListObjects(bucketName, objectPrefix string, recursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo
	ListObjectsV2(bucketName, objectPrefix string, recursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo
}

type MinioUploader interface {
	PutObject(bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (n int64, err error)
	PutObjectWithContext(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (n int64, err error)
	PresignedPutObject(bucketName string, objectName string, expires zone.Duration) (u *url.URL, err error)
	FPutObject(bucketName, objectName, filePath string, opts minio.PutObjectOptions) (n int64, err error)
	FPutObjectWithContext(ctx context.Context, bucketName, objectName, filePath string, opts minio.PutObjectOptions) (n int64, err error)
}

type MinioStater interface {
	StatObject(bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
}
type MinioGetter interface {
	GetObject(bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	GetObjectACL(bucketName, objectName string) (*minio.ObjectInfo, error)
	GetObjectWithContext(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	PresignedGetObject(bucketName string, objectName string, expires zone.Duration, reqParams url.Values) (u *url.URL, err error)
	FGetObject(bucketName, objectName, filePath string, opts minio.GetObjectOptions) error
	FGetObjectWithContext(ctx context.Context, bucketName, objectName, filePath string, opts minio.GetObjectOptions) error
}
