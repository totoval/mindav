package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/totoval/framework/config"
	"github.com/totoval/framework/http/controller"
	"golang.org/x/net/webdav"
	_webdav "totoval/app/logics/webdav"
)

type WebDAV struct {
	controller.BaseController
}

func (wd *WebDAV) Handle (c *gin.Context){
	uri := "/v1" + config.GetString("webdav.base_url")


	h := webdav.Handler{
		Prefix: uri,
		FileSystem: _webdav.FileSystem(),
		LockSystem: webdav.NewMemLS(),
	}

	h.ServeHTTP(c.Writer, c.Request)
	return
}
