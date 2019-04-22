package models

import (
	"math/rand"
	"testing"
	"time"

	"github.com/totoval/framework/helpers/debug"

	"github.com/totoval/framework/helpers/str"

	"totoval/resources/lang"

	"github.com/totoval/framework/database"

	"github.com/totoval/framework/cache"

	"totoval/config"

	"github.com/totoval/framework/helpers/ptr"

	"github.com/totoval/framework/helpers/m"
)

func init() {
	config.Initialize()
	cache.Initialize()
	database.Initialize()
	m.Initialize()
	lang.Initialize() // an translation must contains resources/lang/xx.json file (then a resources/lang/validation_translator/xx.go)
}

func TestUserAffiliation_InsertNode(t *testing.T) {
	// add root user
	root := addUser()
	rootUaff := addAffiliation(root)

	randUaff := rootUaff
	rand.Seed(time.Now().Unix())

	//add invitation
	for i := 0; i < 20; i++ {
		u := addUser()
		var uaff UserAffiliation

		debug.Dump(randUaff.Code)
		if randUaff.Code != nil {
			uaff = addAffiliation(u, *randUaff.Code)
		} else {
			uaff = addAffiliation(u)
		}
		if rand.Intn(10) > 5 {
			randUaff = uaff
		}
	}
}

func TestUserAffiliation_All(t *testing.T) {
	// dump tree
	var uaff UserAffiliation
	debug.Dump(uaff.All())
}

func addUser() User {
	user := User{
		Email:    ptr.String(str.RandString(10) + "@zhigui.com"),
		Password: ptr.String(str.RandString(20)),
	}
	if err := m.H().Create(&user); err != nil {
		panic(err)
	}

	return user
}
func addAffiliation(user User, fromCode ...string) UserAffiliation {
	uaffPtr := &UserAffiliation{
		UserID: user.ID,
	}
	if err := uaffPtr.InsertNode(&user, fromCode...); err != nil {
		panic(err)
	}

	if err := m.H().First(uaffPtr, false); err != nil {
		panic(err)
	}

	return *uaffPtr
}
