package modules

import (
	"log"
	"github.com/eris-ltd/deCerver-interfaces/core"
	"github.com/eris-ltd/deCerver-interfaces/api"
)

type Module interface {
    // for registering handlers to the ate vm
    RegisterModule(registry api.ApiRegistry, logger core.LogSystem) error 

	Init() error
    Start() error
	Shutdown() error

    ReadConfig(config_file string)
    WriteConfig(config_file string)
	Name() string
}

type Database interface {
	Module
    // generalized get/push
	Get(cmd string, params ... string) (string, error)
	Push(cmd string, params ... string) (string, error)
    // get ordered map of storage
    State() core.State
    // ordered map of values in storage (generalized sql table)
    Storage(addr string) core.Storage
    // commit cached data (mine a block)
	Commit()
    // commit continuously
	AutoCommit(toggle bool)
	IsAutocommit() bool
}

