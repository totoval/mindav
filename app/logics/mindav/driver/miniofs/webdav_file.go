package miniofs

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/totoval/framework/helpers/log"
	"github.com/totoval/framework/helpers/toto"
	"io"
	"os"
	"strings"
)

type file struct {
	m *minioFileSystem
	*minio.Object
	name string
}

func (mo file) Stat() (os.FileInfo, error) {
	log.Trace("file stat", toto.V{"name": mo.name})
	return mo.m.Stat(context.Background(), mo.name)
}
func (mo file) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = mo.m.client.PutObject(mo.m.bucketName, strings.TrimPrefix(mo.name, "/"), r, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return 0, log.Error(err, toto.V{"op": "ReadFrom", "name": mo.name})
	}
	fmt.Println("Successfully uploaded bytes: ", n)
	return n, nil
}
func (mo file) Write(p []byte) (n int, err error) {
	return len(p), nil // useless
}

func (mo file) Readdir(count int) (fileInfoList []os.FileInfo, err error) {
	log.Trace("file readDir", toto.V{"name": mo.name})

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
