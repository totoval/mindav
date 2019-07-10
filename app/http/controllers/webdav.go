package controllers

import (
	"github.com/totoval/framework/http/controller"
	"github.com/totoval/framework/request"
	"totoval/app/logics/mindav"
)

type WebDAV struct {
	controller.BaseController
}
func (wd *WebDAV) Handle (c *request.Context){
	mindav.Handler().ServeHTTP(c.Writer, c.Request)
	return
}
