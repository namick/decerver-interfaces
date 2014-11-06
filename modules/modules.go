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
	// generalized get/push
	Get(cmd string, params ...string) (string, error)
	Push(cmd string, params ...string) (string, error)

	GetState() State
	GetStorage(target string) Storage
    GetStorageAt(target, storage string) string
    
    Tx(addr, amt string) // TODO: return hash
    Msg(addr string, data []string) // TODO: return hash
    Script(file, lang string) string // TODO: remove lang

    // TODO: allow set gas/price/amts
    
    // subscribe to event
    Subscribe(name, event, target string) chan events.Event

	// commit cached data (mine a block)
	Commit()
	// commit continuously
	AutoCommit(toggle bool)
	IsAutocommit() bool
}

type FileSystem interface {
    Get(cmd string, params ...string) (interface{}, error)
    Push(cmd string, params ...string) (string, error)

    GetBlock(hash string) ([]byte, error)
    GetFile(hash string) ([]byte, error)
    GetStream(hash string) (chan []byte, error)
    GetTree(hash string, depth int) (FsNode, error)

    PushBlock(block []byte) (string, error)
    PushBlockString(block string) (string, error)
    PushFile(fpath string) (string, error)
    PushTree(fpath string, depth int) (string, error)

    Subscribe(name string, event string, target string) chan events.Event
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

// File System Node for directory trees
type FsNode struct{
    Nodes []*FsNode
    Name string
    Hash string
}
