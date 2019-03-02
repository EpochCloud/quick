package app

import(
	"quick/config"
)


var (
	errResult    	= config.NewErrorResult()
	Manager      	= config.NewManagers()
	succeedResult   = config.NewResult()
	Service         = config.NewManagers()
	DeleteService	= config.NewManagers()
	InsertService    = config.NewManagers()
)