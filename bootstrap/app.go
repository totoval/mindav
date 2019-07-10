package bootstrap

import (
	"github.com/totoval/framework/cache"
	"github.com/totoval/framework/helpers/zone"
	"github.com/totoval/framework/logs"
	"github.com/totoval/framework/sentry"
	"github.com/totoval/framework/validator"
	"totoval/app/logics/mindav"

	"totoval/config"
	"totoval/resources/lang"
)

func Initialize() {
	config.Initialize()
	sentry.Initialize()
	logs.Initialize()
	zone.Initialize()
	lang.Initialize() // an translation must contains resources/lang/xx.json file (then a resources/lang/validation_translator/xx.go)
	cache.Initialize()
	// database.Initialize()
	// m.Initialize()
	// queue.Initialize()
	// jobs.Initialize()
	// events.Initialize()
	// listeners.Initialize()
	mindav.Initialize()

	validator.UpgradeValidatorV8toV9()
}
