package mindav

import (
	"github.com/totoval/framework/config"
	"golang.org/x/net/webdav"
)

var mindavHandler webdav.Handler

func Initialize() {

	uri := "/v1" + config.GetString("webdav.base_url")
	mindavHandler = webdav.Handler{
		Prefix:     uri,
		FileSystem: fileSystem(),
		LockSystem: webdav.NewMemLS(),
	}
}

func Handler() *webdav.Handler {
	return &mindavHandler
}
