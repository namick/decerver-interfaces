package legalmarkdown

import (
	// "github.com/eris-ltd/legalmarkdown"
	"github.com/eris-ltd/decerver-interfaces/core"
	"github.com/eris-ltd/decerver-interfaces/events"
)

type LmdApi struct {
	
}

// TODO api funcs.
func (lmda * LmdApi) DoSomething(){
	
}

// implements decerver-interface module
type LmdModule struct {
	api *LmdApi
}

func NewLmdModule() *LmdModule {
	lmdApi := &LmdApi{}
	return &LmdModule{lmdApi}
}

func (mod *LmdModule) Register(fileIO core.FileIO, rm core.RuntimeManager, eReg events.EventRegistry) error {
	// rm.RegisterApiObject("lmd", mod.api)
	return nil
}

func (mod *LmdModule) Init() error {
	
	return nil
}

func (mod *LmdModule) Start() error {
	return nil
}

// TODO: UDP socket won't close
// https://github.com/jbenet/go-ipfs/issues/389
func (mod *LmdModule) Shutdown() error {
	
	return nil
}

func (mod *LmdModule) Restart() error {
	err := mod.Shutdown()
	if err != nil {
		return nil
	}
	return mod.Start();
}

func (mod *LmdModule) SetProperty(name string, data interface{}) {
}

func (mod *LmdModule) Property(name string) interface{} {
	return nil
}

func (mod *LmdModule) ReadConfig(config_file string) {
}

func (mod *LmdModule) WriteConfig(config_file string) {
}

func (mod *LmdModule) Name() string {
	return "lmd"
}

func (mod *LmdModule) Subscribe(name string, event string, target string) chan events.Event {
	return nil
}

func (mod *LmdModule) UnSubscribe(name string) {
}