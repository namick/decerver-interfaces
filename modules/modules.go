package modules

import (
	"log"
	"github.com/eris-ltd/deCerver-interfaces/core"
	"github.com/eris-ltd/deCerver-interfaces/api"
)

type Module interface {
	// How to get the active logger
	Logger() *log.Logger
	
	Init(se core.ScriptEngine)
	Name() string
	HttpAPIServices() []interface{}
	WsAPIServiceFactories() []api.WsAPIServiceFactory
	Shutdown()
}

type Database interface {
	Module
	Get(addr string, params ... string)
	Push(addr string, params ... string)
	Commit()
	AutoCommit(toggle bool)
	IsAutocommit() bool
}

