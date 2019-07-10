package mindav

import (
	"github.com/totoval/framework/config"
	"golang.org/x/net/webdav"
	"totoval/app/logics/mindav/driver"
	//"totoval/app/logics/mindav/driver/minio"
)

func fileSystem(namespace string) webdav.FileSystem {
	switch config.GetString("webdav.driver") {
	case "memory":
		return webdav.NewMemFS()
	case "file":
		return webdav.Dir(config.GetString("webdav.filesystems.file.base_path"))
	case "minio":
		//return minio.New(namespace, config.GetString("webdav.filesystems.minio.location"))
		return driver.NewMinio(namespace, config.GetString("webdav.filesystems.minio.location"))
	default:
		panic("please set a filesystem in the config")
	}
}


