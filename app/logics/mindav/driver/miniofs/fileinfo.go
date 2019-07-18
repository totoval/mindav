package miniofs

import (
	"github.com/minio/minio-go/v6"
	"github.com/totoval/framework/helpers/log"
	"github.com/totoval/framework/helpers/toto"
	"github.com/totoval/framework/helpers/zone"
	"os"
	"path"
	"strings"
)

type fileInfo struct {
	minio.ObjectInfo
}

func (moi fileInfo) Name() string {
	name := moi.ObjectInfo.Key

	name = strings.Trim(name, "/")

	if strings.Contains(name, "/") {
		name = path.Clean(strings.Replace(name, path.Dir(name), "", 1))
	}
	log.Debug(name, toto.V{"Key": moi.ObjectInfo.Key, "ObjectName": name})

	return name
} // base name of the file
func (moi fileInfo) Size() int64 {
	return moi.ObjectInfo.Size
} // length in bytes for regular files; system-dependent for others
func (moi fileInfo) Mode() os.FileMode {
	return 777
} // file mode bits
func (moi fileInfo) ModTime() zone.Time {
	return moi.ObjectInfo.LastModified
} // modification time
func (moi fileInfo) IsDir() bool {
	isDir := moi.ObjectInfo.ContentType == "inode/directory"
	log.Debug(moi.ObjectInfo.Key, toto.V{"IsDir": isDir})
	return isDir
} // abbreviation for Mode().IsDir()
func (moi fileInfo) Sys() interface{} {
	return nil
} // underlying data source (can return nil)
