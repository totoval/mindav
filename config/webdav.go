package config

import (
	. "github.com/totoval/framework/config"
)

func init() {
	webdav := make(map[string]interface{})

	webdav["filesystem"] = "memory" // file, memory

	webdav["base_path"] = "." // for "file" filesystem
	webdav["supported_folder_depth"] = 10
	webdav["base_url"] = "/webdav"

	Add("webdav", webdav)
}
