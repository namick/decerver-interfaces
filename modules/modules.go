package modules

import (
	"github.com/eris-ltd/deCerver-interfaces/api"
	"github.com/eris-ltd/deCerver-interfaces/core"
	"github.com/eris-ltd/deCerver-interfaces/events"
)

type Module interface {
	// For registering with deCerver.
	Register(fileIO core.FileIO, registry api.ApiRegistry, runtime core.Runtime, eReg events.EventRegistry) error
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
	Get(cmd string, params ...string) (string, error)
	Push(cmd string, params ...string) (string, error)
	// get ordered map of storage
	State() State
	// ordered map of values in storage (generalized sql table)
	Storage(target string) Storage
	// commit cached data (mine a block)
	Commit()
	// commit continuously
	AutoCommit(toggle bool)
	IsAutocommit() bool
}

// TODO implement this
type FileSystem interface {
	GetFile(hash string)
}

// Ordered map for storage in an account or generalized table
type Storage struct {
	// hex strings for eth, arrays of strings (cols) for sql dbs
	Storage map[string]interface{}
	Order   []string
}

// Ordered map for all accounts
type State struct {
	State map[string]Storage // map addrs to map of storage to value
	Order []string           // ordered addrs and ordered storage inside
}
