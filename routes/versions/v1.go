package versions

import (
	"github.com/totoval/framework/config"
	"github.com/totoval/framework/request"
	"github.com/totoval/framework/route"
	"totoval/routes/groups"
)

func NewV1(engine *request.Engine) {
	ver := route.NewVersion(engine, "v1") // here must have a version

	// auth routes
	// ver.Auth("", func(grp route.Grouper) {
	// 	grp.AddGroup("/user", &groups.UserGroup{})
	// })

	// no auth routes
	ver.NoAuth("", func(grp route.Grouper) {
		grp.AddGroup(config.GetString("webdav.base_url"), &groups.WebDAVGroup{})
		// grp.AddGroup("", &groups.AuthGroup{})
		// grp.AddGroup("/user-affiliation", &groups.UserAffiliationGroup{})
	})
}
