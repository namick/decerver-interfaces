package modules

import (
	"fmt"
	"github.com/eris-ltd/decerver-interfaces/core"
	"github.com/eris-ltd/decerver-interfaces/events"
	"github.com/eris-ltd/decerver-interfaces/types"
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
		Restart() error
		Shutdown() error
		Name() string
		// TODO No channel here. Wait so that modules doesn't break.
		Subscribe(name, event, target string) chan events.Event
		UnSubscribe(name string)

		SetProperty(name string, data interface{})
		Property(name string) interface{}
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
	State() JsObject
	Storage(target string) JsObject
	Account(target string) JsObject
	StorageAt(target, storage string) JsObject

	BlockCount() JsObject
	LatestBlock() JsObject
	Block(hash string) JsObject

	IsScript(target string) JsObject

	Tx(addr, amt string) JsObject
	Msg(addr string, data []string) JsObject
	Script(file, lang string) JsObject

	// TODO: allow set gas/price/amts
	// subscribe to event

	// commit cached txs (mine a block)
	Commit() JsObject
	// commit continuously
	AutoCommit(toggle bool) JsObject
	IsAutocommit() JsObject
}

type KeyManager interface {
	ActiveAddress() JsObject
	Address(n int) JsObject
	SetAddress(addr string) JsObject
	SetAddressN(n int) JsObject
	NewAddress(set bool) JsObject
	// Don't want to pass numbers from otto if it can be avoided
	// (otto tends to switch around between int and float types).
	Addresses() JsObject
	AddressCount() JsObject
}

// Default JsObjects comes with the data + an error field, like this:
// Data is a string
// {
//	   "Data" : data,
//	   "Error" : error
// }

type FileSystem interface {
	KeyManager

	Get(cmd string, params ...string) JsObject
	Push(cmd string, params ...string) JsObject // string

	GetBlock(hash string) JsObject           // []byte
	GetFile(hash string) JsObject            // []byte
	GetStream(hash string) JsObject          // []byte
	GetTree(hash string, depth int) JsObject // FsNode

	PushBlock(block []byte) JsObject           // string
	PushBlockString(block string) JsObject     // string
	PushFile(fpath string) JsObject            // string
	PushTree(fpath string, depth int) JsObject // string
}

type Compiler interface {
	Compile(interface{}) JsObject
}

// Converts a data and an error value into a javascript ready object.
// All methods on objects that modules bind to the js runtime should return
// this.
func JsReturnVal(data interface{}, err error) JsObject {
	ret := make(JsObject)
	if err != nil {
		ret["Error"] = err.Error()
		ret["Data"] = nil
	} else {
		ret["Error"] = ""
		ret["Data"] = types.ToJsValue(data)
	}
	fmt.Printf("JS RETURN VALUE: %v\n",ret)
	return ret
}

// If there is no error
func JsReturnValNoErr(data interface{}) JsObject {
	ret := make(JsObject)
	ret["Error"] = ""
	ret["Data"] = types.ToJsValue(data)
	return ret
}

// If there is only an error
func JsReturnValErr(err error) JsObject {
	ret := make(JsObject)
	ret["Error"] = err.Error()
	ret["Data"] = nil
	return ret
}