package groups

import (
	"fmt"
	"github.com/totoval/framework/config"
	"github.com/totoval/framework/request"
	"github.com/totoval/framework/route"

	"totoval/app/http/controllers"
)

type WebDAVGroup struct {
	WebDAVController controllers.WebDAV
}
func (wdg *WebDAVGroup)Group(group route.Grouper) {
	webDAVGroup := WebDAVRouterGroup{group}
	// webDAVGroup.WebDAVAny(config.GetString("webdav.base_url"), wdg.WebDAVController.Handle)
	// webDAVGroup.WebDAVAny(config.GetString("webdav.base_url") + "/", wdg.WebDAVController.Handle)
	//
	// webDAVGroup.WebDAVAnyPaths(config.GetString("webdav.base_url"), wdg.WebDAVController.Handle)
	webDAVGroup.WebDAVAny("", wdg.WebDAVController.Handle)
	webDAVGroup.WebDAVAny("/", wdg.WebDAVController.Handle)

	webDAVGroup.WebDAVAnyPaths("", wdg.WebDAVController.Handle)
}

type WebDAVRouterGroup struct {
	route.Grouper
}
func (wdrg *WebDAVRouterGroup) WebDAVAny(relativePath string, handlers ...request.HandlerFunc) {
	wdrg.Any(relativePath, handlers...)
	wdrg.Handle("PROPFIND", relativePath, handlers...)
	wdrg.Handle("PROPPATCH", relativePath, handlers...)
	wdrg.Handle("MKCOL", relativePath, handlers...)
	wdrg.Handle("COPY", relativePath, handlers...)
	wdrg.Handle("MOVE", relativePath, handlers...)
	wdrg.Handle("LOCK", relativePath, handlers...)
	wdrg.Handle("UNLOCK", relativePath, handlers...)
}
func (wdrg *WebDAVRouterGroup) WebDAVAnyPaths(basePath string, handlers ...request.HandlerFunc) {
	currentPath := "/"
	for i:=0; i<config.GetInt("webdav.supported_folder_depth"); i++ {
		p := fmt.Sprintf(":path%d", i)
		currentPath += p
		wdrg.WebDAVAny(basePath + currentPath, handlers...)
		currentPath += "/"
		wdrg.WebDAVAny(basePath + currentPath, handlers...)
	}
}