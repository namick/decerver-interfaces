package monk

import (
	"github.com/eris-ltd/deCerver-interfaces/core"
	"github.com/eris-ltd/deCerver-interfaces/api"
	"github.com/eris-ltd/thelonious/monk"
	"github.com/golang/glog"
)

type MonkModule struct {
	ethChain *monk.EthChain
	wsAPIServiceFactory api.WsAPIServiceFactory
	httpAPIService interface{}
	
}

func NewMonkModule() *MonkModule{
	return &MonkModule{}
}

func (mm *MonkModule) Register(fileIO core.FileIO, registry api.ApiRegistry, runtime core.Runtime) error {
	
	//logSystem.AddLogger(logger)
	// Monk ethchain
	mm.ethChain = monk.NewEth(nil)
	mm.ethChain.Init()
	mm.ethChain.Start()
	// The json-rpc service
	httpAPI := &Monk{}
	httpAPI.EthChain = mm.ethChain 
	mm.httpAPIService = httpAPI
	registry.RegisterHttpServices(httpAPI)
	
	fact := NewMonkWsAPIFactory(mm.ethChain)
	mm.wsAPIServiceFactory = fact
	registry.RegisterWsServiceFactories(fact)
	
	return nil
}

func (mm *MonkModule) Init() error {
	return nil
}

func (mm *MonkModule) Start() error {
	return nil
}

func (mm *MonkModule) ReadConfig(config_file string) {
	mm.ethChain.ReadConfig(config_file)
}

func (mm *MonkModule) WriteConfig(config_file string){
	
}

func (mm *MonkModule) Name() string {
	return "MonkModule"
}

func (mm *MonkModule) Shutdown() error {
	glog.Infoln("Goodbye from MonkModule")
	return nil
}