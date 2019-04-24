package groups

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/totoval/framework/config"

	"totoval/app/http/controllers"
)

type WebDAVGroup struct {
	WebDAVController controllers.WebDAV
}
type WebDAVRouterGroup struct {
	*gin.RouterGroup
}
func (wdrg *WebDAVRouterGroup) WebDAVAny(relativePath string, handlers ...gin.HandlerFunc) {
	wdrg.Any(relativePath, handlers...)
	wdrg.Handle("PROPFIND", relativePath, handlers...)
	wdrg.Handle("PROPPATCH", relativePath, handlers...)
	wdrg.Handle("MKCOL", relativePath, handlers...)
	wdrg.Handle("COPY", relativePath, handlers...)
	wdrg.Handle("MOVE", relativePath, handlers...)
	wdrg.Handle("LOCK", relativePath, handlers...)
	wdrg.Handle("UNLOCK", relativePath, handlers...)
}
func (wdrg *WebDAVRouterGroup) WebDAVAnyPaths(basePath string, handlers ...gin.HandlerFunc) {
	currentPath := "/"
	for i:=0; i<config.GetInt("webdav.supported_folder_depth"); i++ {
		p := fmt.Sprintf(":path%d", i)
		currentPath += p
		wdrg.WebDAVAny(basePath + currentPath, handlers...)
		currentPath += "/"
		wdrg.WebDAVAny(basePath + currentPath, handlers...)
	}
}

func (wdg *WebDAVGroup) Register(group *gin.RouterGroup) {

	newGroup := group.Group("")
	{
		webDAVGroup := WebDAVRouterGroup{newGroup}
		webDAVGroup.WebDAVAny(config.GetString("webdav.base_url"), wdg.WebDAVController.Handle)
		webDAVGroup.WebDAVAny(config.GetString("webdav.base_url") + "/", wdg.WebDAVController.Handle)

		webDAVGroup.WebDAVAnyPaths(config.GetString("webdav.base_url"), wdg.WebDAVController.Handle)
	}
}
