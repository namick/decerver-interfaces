package monkjs

import (
	"fmt"
	"github.com/eris-ltd/decerver-interfaces/core"
	"github.com/eris-ltd/decerver-interfaces/events"
	"github.com/eris-ltd/decerver-interfaces/modules"
	"github.com/eris-ltd/thelonious/monk"
)

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

func (mjs *MonkJs) WorldState() modules.JsObject {
	ws := mjs.mm.WorldState()
	return modules.JsReturnVal(modules.ToMap(ws), nil)
}

func (mjs *MonkJs) State() modules.JsObject {
	return modules.JsReturnVal(modules.ToMap(mjs.mm.State()), nil)
}

func (mjs *MonkJs) Storage(target string) modules.JsObject {
	return modules.JsReturnVal(modules.ToMap(mjs.mm.Storage(target)), nil)
}

func (mjs *MonkJs) Account(target string) modules.JsObject {
	return modules.JsReturnVal(modules.ToMap(mjs.mm.Account(target)), nil)
}

func (mjs *MonkJs) StorageAt(target, storage string) modules.JsObject {
	return modules.JsReturnVal(mjs.mm.StorageAt(target, storage), nil)
}

func (mjs *MonkJs) BlockCount() modules.JsObject {
	return modules.JsReturnVal(mjs.mm.BlockCount(), nil)
}

func (mjs *MonkJs) LatestBlock() modules.JsObject {
	return modules.JsReturnVal(mjs.mm.LatestBlock(), nil)
}

func (mjs *MonkJs) Block(hash string) modules.JsObject {
	return modules.JsReturnVal(modules.ToMap(mjs.mm.Block(hash)), nil)
}

func (mjs *MonkJs) IsScript(target string) modules.JsObject {
	return modules.JsReturnVal(mjs.mm.IsScript(target), nil)
}

func (mjs *MonkJs) Tx(addr, amt string) modules.JsObject {
	hash, err := mjs.mm.Tx(addr, amt)
	var ret modules.JsObject
	if err == nil {
		ret = make(modules.JsObject)
		ret["Hash"] = hash
		ret["Address"] = ""
		ret["Error"] = ""
	}
	return modules.JsReturnVal(ret, err)
}

func (mjs *MonkJs) Msg(addr string, data []interface{}) modules.JsObject {
	indata := make([]string, 0)
	for _, d := range data {
		str, ok := d.(string)
		if !ok {
			return modules.JsReturnValErr(fmt.Errorf("Msg indata is not an array of strings"))
		}
		indata = append(indata, str)
	}
	hash, err := mjs.mm.Msg(addr, indata)
	var ret modules.JsObject
	if err == nil {
		ret = make(modules.JsObject)
		ret["Hash"] = hash
		ret["Address"] = ""
		ret["Error"] = ""
	}
	return modules.JsReturnVal(ret, err)
}

func (mjs *MonkJs) Script(file, lang string) modules.JsObject {
	addr, err := mjs.mm.Script(file, lang)
	var ret modules.JsObject
	if err == nil {
		ret = make(modules.JsObject)
		ret["Hash"] = ""
		ret["Address"] = addr
		ret["Error"] = ""
	}
	return modules.JsReturnVal(ret, err)
}

func (mjs *MonkJs) Commit() modules.JsObject {
	mjs.mm.Commit()
	return modules.JsReturnVal(nil, nil)
}

func (mjs *MonkJs) AutoCommit(toggle bool) modules.JsObject {
	mjs.mm.AutoCommit(toggle)
	return modules.JsReturnVal(nil, nil)
}

func (mjs *MonkJs) IsAutocommit() modules.JsObject {
	return modules.JsReturnVal(mjs.mm.IsAutocommit(), nil)
}

/*
   Module should also satisfy KeyManager
*/

func (mjs *MonkJs) ActiveAddress() modules.JsObject {
	return modules.JsReturnVal(mjs.mm.ActiveAddress(), nil)
}

func (mjs *MonkJs) Addresses() modules.JsObject {
	count := mjs.mm.AddressCount()
	addresses := make(modules.JsObject)
	array := make([]string, count)

	for i := 0; i < count; i++ {
		addr, _ := mjs.mm.Address(i)
		array[i] = addr
	}
	addresses["Addresses"] = array
	return modules.JsReturnVal(addresses, nil)
}

func (mjs *MonkJs) SetAddress(addr string) modules.JsObject {
	err := mjs.mm.SetAddress(addr)
	if err != nil {
		return modules.JsReturnValErr(err)
	} else {
		// No error means success.
		return modules.JsReturnValNoErr(nil)
	}
}

// TODO Not used atm. Think about this.
func (mjs *MonkJs) SetAddressN(n int) modules.JsObject {
	mjs.mm.SetAddressN(n)
	return modules.JsReturnValNoErr(nil)
}

func (mjs *MonkJs) NewAddress(set bool) modules.JsObject {
	return modules.JsReturnValNoErr(mjs.mm.NewAddress(set))
}

func (mjs *MonkJs) AddressCount() modules.JsObject {
	return modules.JsReturnValNoErr(mjs.mm.AddressCount())
}
