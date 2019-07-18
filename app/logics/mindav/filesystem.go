package mindav

import (
	"github.com/totoval/framework/config"
	"golang.org/x/net/webdav"
	"totoval/app/logics/mindav/driver/miniofs"

	// "totoval/app/logics/mindav/driver/miniofs"
)

func fileSystem() webdav.FileSystem {
	switch config.GetString("webdav.driver") {
	case "memory":
		return webdav.NewMemFS()
	case "file":
		return webdav.Dir(config.GetString("webdav.filesystems.file.base_path"))
	case "minio":
		return  miniofs.New(config.GetString("webdav.filesystems.minio.bucket"), config.GetString("webdav.filesystems.minio.location"))
	default:
		panic("please set a filesystem in the config")
	}
}


