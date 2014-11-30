package modules

import (
	"github.com/eris-ltd/decerver-interfaces/core"
	"github.com/eris-ltd/decerver-interfaces/events"
)

type JsObject map[string]interface{}

type (
	
	ModuleInfo struct {
		Name       string      `json:"name"`
		Version    string      `json:"version"`
		Author     *AuthorInfo `json:"author"`
		Licence    string      `json:"licence"`
		Repository string      `json:"repository"`
	}

	AuthorInfo struct {
		Name  string `json:"name"`
		EMail string `json:"e-mail"`
	}
)

type (
	Module interface {
		// For registering with decerver.
		Register(fileIO core.FileIO, rm core.RuntimeManager, eReg events.EventRegistry) error
		Init() error
		Start() error
		Shutdown() error
		Name() string
		Subscribe(name, event, target string) chan events.Event
		UnSubscribe(name string)
	}

	ModuleRegistry interface {
		GetModules() map[string]Module
		GetModuleNames() []string
	}
)

// TODO: interface for history (transactions, transaction pool)
type Blockchain interface {
	KeyManager
	WorldState() JsObject
	State() *State
	Storage(target string) JsObject
	Account(target string) JsObject
	StorageAt(target, storage string) string

	BlockCount() int
	LatestBlock() string
	Block(hash string) JsObject

	IsScript(target string) bool

	Tx(addr, amt string) (string, error)
	Msg(addr string, data []string) JsObject
	Script(file, lang string) JsObject

	// TODO: allow set gas/price/amts
	// subscribe to event

	// commit cached txs (mine a block)
	Commit()
	// commit continuously
	AutoCommit(toggle bool)
	IsAutocommit() bool
}

type KeyManager interface {
	ActiveAddress() string
	Address(n int) JsObject
	SetAddress(addr string) string
	SetAddressN(n int) string
	NewAddress(set bool) string
	// Don't want to pass numbers from otto if it can be avoided.
	Addresses() JsObject
	AddressCount() int
}

type FileSystem interface {
	KeyManager

	Get(cmd string, params ...string) (interface{}, error)
	Push(cmd string, params ...string) (string, error)

	GetBlock(hash string) ([]byte, error)
	GetFile(hash string) ([]byte, error)
	GetStream(hash string) (chan []byte, error)
	GetTree(hash string, depth int) (*FsNode, error)

	PushBlock(block []byte) (string, error)
	PushBlockString(block string) (string, error)
	PushFile(fpath string) (string, error)
	PushTree(fpath string, depth int) (string, error)

	Subscribe(name string, event string, target string) chan events.Event
	UnSubscribe(name string)
}
