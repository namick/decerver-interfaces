package monkjs

import (
	"github.com/eris-ltd/decerver-interfaces/core"
	"github.com/eris-ltd/decerver-interfaces/events"
	"github.com/eris-ltd/decerver-interfaces/modules"
	"github.com/eris-ltd/thelonious/monk"
	//"fmt"
)

// Typeless map used to represent everything that is passed to front- and back-end javascript.
type JsObject map[string]interface{}

// implements decerver-interfaces Module
type MonkJs struct {
	mm *monk.MonkModule
}

func NewMonkJs() *MonkJs {
	monkModule := monk.NewMonk(nil)
	return &MonkJs{monkModule}
}

// register the module with the decerver javascript vm
func (mjs *MonkJs) Register(fileIO core.FileIO, rm core.RuntimeManager, eReg events.EventRegistry) error {
	rm.RegisterApi("monk", mjs)
	return nil
}

// initialize an monkchain
// it may or may not already have an ethereum instance
// basically gives you a pipe, local keyMang, and reactor
func (mjs *MonkJs) Init() error {
	return mjs.mm.Init()
}

// start the ethereum node
func (mjs *MonkJs) Start() error {
	return mjs.mm.Start()
}

func (mjs *MonkJs) Shutdown() error {
	return mjs.mm.Shutdown()
}

// ReadConfig and WriteConfig implemented in config.go

// What module is this?
func (mjs *MonkJs) Name() string {
	return "monk"
}

func (mjs *MonkJs) Subscribe(name, event, target string) chan events.Event {
	return mjs.mm.Subscribe(name, event, target)
}

func (mjs *MonkJs) UnSubscribe(name string) {
	mjs.mm.UnSubscribe(name)
}

/*
   Wrapper so module satisfies Blockchain
*/

func (mjs *MonkJs) WorldState() JsObject {
	ws := mjs.mm.WorldState()
	return modules.ToMap(ws)
}

func (mjs *MonkJs) State() JsObject {
	return modules.ToMap(mjs.mm.State())
}

func (mjs *MonkJs) Storage(target string) JsObject {
	return modules.ToMap(mjs.mm.Storage(target))
}

func (mjs *MonkJs) Account(target string) JsObject {
	return modules.ToMap(mjs.mm.Account(target))
}

func (mjs *MonkJs) StorageAt(target, storage string) string {
	return mjs.mm.StorageAt(target, storage)
}

func (mjs *MonkJs) BlockCount() int {
	return mjs.mm.BlockCount()
}

func (mjs *MonkJs) LatestBlock() string {
	return mjs.mm.LatestBlock()
}

func (mjs *MonkJs) Block(hash string) JsObject {
	return modules.ToMap(mjs.mm.Block(hash))
}

func (mjs *MonkJs) IsScript(target string) bool {
	return mjs.mm.IsScript(target)
}

func (mjs *MonkJs) Tx(addr, amt string) JsObject {
	hash, err := mjs.mm.Tx(addr, amt)
	ret := make(JsObject)
	if err != nil {
		ret["Hash"] = ""
		ret["Address"] = ""
		ret["Error"] = err.Error()
	} else {
		ret["Hash"] = hash
		ret["Address"] = ""
		ret["Error"] = ""
	}
	return ret
}

func (mjs *MonkJs) Msg(addr string, data []string) JsObject {
	hash, err := mjs.mm.Msg(addr, data)
	ret := make(JsObject)
	if err != nil {
		ret["Hash"] = ""
		ret["Address"] = ""
		ret["Error"] = err.Error()
	} else {
		ret["Hash"] = hash
		ret["Address"] = ""
		ret["Error"] = ""
	}
	return ret
}

func (mjs *MonkJs) Script(file, lang string) JsObject {
	addr, err := mjs.mm.Script(file, lang)
	ret := make(JsObject)
	if err != nil {
		ret["Hash"] = ""
		ret["Address"] = ""
		ret["Error"] = err.Error()
	} else {
		ret["Hash"] = ""
		ret["Address"] = addr
		ret["Error"] = ""
	}
	return ret
}

func (mjs *MonkJs) Commit() {
	mjs.mm.Commit()
}

func (mjs *MonkJs) AutoCommit(toggle bool) {
	mjs.mm.AutoCommit(toggle)
}

func (mjs *MonkJs) IsAutocommit() bool {
	return mjs.mm.IsAutocommit()
}

/*
   Module should also satisfy KeyManager
*/

func (mjs *MonkJs) ActiveAddress() string {
	return mjs.mm.ActiveAddress()
}

func (mjs *MonkJs) Addresses() JsObject {
	count := mjs.AddressCount()
	addresses := make(JsObject)
	array := make([]string, count)
	
	for i := 0; i < count; i++ {
		addr, _ := mjs.mm.Address(i)
		array[i] = addr
	}
	addresses["Addresses"] = array
	return addresses
}

func (mjs *MonkJs) SetAddress(addr string) bool {
	err := mjs.mm.SetAddress(addr)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (mjs *MonkJs) SetAddressN(n int) error {
	return mjs.mm.SetAddressN(n)
}

func (mjs *MonkJs) NewAddress(set bool) string {
	return mjs.mm.NewAddress(set)
}

func (mjs *MonkJs) AddressCount() int {
	return mjs.mm.AddressCount()
}
