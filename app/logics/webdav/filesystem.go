package _webdav

import (
	"github.com/totoval/framework/config"
	"golang.org/x/net/webdav"
)

func FileSystem() webdav.FileSystem {
	switch config.GetString("webdav.filesystem") {
	case "memory":
		return webdav.NewMemFS()
	case "file":
		return webdav.Dir(config.GetString("webdav.base_path"))
	default:
		panic("please set a filesystem in the config")
	}
}