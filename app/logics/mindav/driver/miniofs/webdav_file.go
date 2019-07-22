package miniofs

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/totoval/framework/config"
	"github.com/totoval/framework/helpers/debug"
	"github.com/totoval/framework/helpers/hash"
	"github.com/totoval/framework/helpers/log"
	"github.com/totoval/framework/helpers/toto"
	"io"
	"os"
	"path"
	"strings"
)

type file struct {
	m *minioFileSystem
	*minio.Object
	name string
}

func (mo *file) Stat() (os.FileInfo, error) {
	log.Trace("file stat", toto.V{"name": mo.name})
	return mo.m.Stat(context.Background(), mo.name)
}
func (mo *file) ReadFrom(r io.Reader) (n int64, err error) {

	// memory mode
	if config.GetBool("webdav.memory_upload_mode") {
		n, err = mo.m.client.PutObject(mo.m.bucketName, strings.TrimPrefix(mo.name, "/"), r, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err != nil {
			return 0, log.Error(err, toto.V{"op": "ReadFrom", "name": mo.name})
		}
		fmt.Println("Successfully uploaded bytes: ", n)
		return n, nil
	}

	// file mode
	tmpFilePath := path.Join(mo.m.uploadTmpPath, hash.Md5(mo.name))
	f, err := os.Create(tmpFilePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	defer func(p string) {
		err = os.RemoveAll(p)
		if err != nil {
			_ = log.Error(err, toto.V{"op": "upload", "name": mo.name, "tempName": p})
		}
	}(tmpFilePath)

	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := f.Write(buf[:n]); err != nil {
			return 0, err
		}
	}
	n, err = mo.m.client.FPutObject(mo.m.bucketName, strings.TrimPrefix(mo.name, "/"), tmpFilePath, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return 0, log.Error(err, toto.V{"op": "ReadFrom", "name": mo.name})
	}

	log.Trace(hash.Md5(mo.name), toto.V{"op": "upload", "name": mo.name})

	fmt.Println("Successfully uploaded bytes: ", n)
	return n, nil
}
func (mo *file) Write(p []byte) (n int, err error) {
	debug.DD(p)
	return len(p), nil // useless
}

func (mo *file) Readdir(count int) (fileInfoList []os.FileInfo, err error) {
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

		fileInfoList = append(fileInfoList, &fileInfo{object})
	}

	return fileInfoList, err
}
