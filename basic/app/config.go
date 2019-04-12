package app

import (
	"github.com/DeanThompson/syncmap"
	"quick/config"
)

var (
	errResult     = config.NewErrorResult()
	Manager       = config.NewManagers()
	succeedResult = config.NewResult()
	Service       = config.NewManagers()
	DeleteService = config.NewManagers()
	InsertService = config.NewManagers()
	OperationList = syncmap.New()
)
