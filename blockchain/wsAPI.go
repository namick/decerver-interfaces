package blockchain

// This handles socket-based rpc. Part of it is reacting to requests sent from the
// client, and part of it is reacting to changes in the ethereum world state,
// and propagating these.
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eris-ltd/deCerver-interfaces/api"
	"github.com/eris-ltd/deCerver-interfaces/modules"
	"github.com/eris-ltd/deCerver-interfaces/events"
	"time"
	"github.com/eris-ltd/thelonious/monklog"
	"io"
	"sync"
	"log"
	"bufio"
)

type WebSocketAPIFactory struct {
	bc modules.Blockchain
	ethLogger *EthLogger
	serviceName string
}

func NewWebSocketAPIFactory(bc modules.Blockchain) *WebSocketAPIFactory {
	fact := &WebSocketAPIFactory{
		bc: bc,
		ethLogger:   NewEthLogger(),
		serviceName: "MonkWsAPI",
	}
	return fact
}

func (fact *WebSocketAPIFactory) Init() {

}

func (fact *WebSocketAPIFactory) Shutdown() {
	
}

func (fact *WebSocketAPIFactory) ServiceName() string {
	return fact.serviceName
}

func (fact *WebSocketAPIFactory) CreateService() api.WsAPIService {
	
	service := newWebSocketAPI(fact.bc)
	service.name = fact.serviceName
	service.ethLogger = fact.ethLogger
	return service
}

type WebSocketAPI struct {
	name        string
	mappings    map[string]api.WsAPIMethod
	bc			modules.Blockchain
	conn        api.WebSocketObj
	ethListener *EthListener
	ethLogger   *EthLogger
}

// Create a new handler
func newWebSocketAPI(bc modules.Blockchain) *WebSocketAPI {
	bcAPI := &WebSocketAPI{}
	bcAPI.bc = bc

	bcAPI.mappings = make(map[string]api.WsAPIMethod)
	bcAPI.mappings["MyBalance"] = bcAPI.MyBalance
	bcAPI.mappings["MyAddress"] = bcAPI.MyAddress
	bcAPI.mappings["StartMining"] = bcAPI.StartMining
	bcAPI.mappings["StopMining"] = bcAPI.StopMining
	bcAPI.mappings["LastBlockNumber"] = bcAPI.LastBlockNumber
	bcAPI.mappings["BlockMiniByHash"] = bcAPI.BlockMiniByHash
	bcAPI.mappings["BlockByHash"] = bcAPI.BlockByHash
	bcAPI.mappings["Account"] = bcAPI.Account
	bcAPI.mappings["Transact"] = bcAPI.Transact
	bcAPI.mappings["WorldState"] = bcAPI.WorldState

	return bcAPI
}

func (bcAPI *WebSocketAPI) SetConnection(wsConn api.WebSocketObj) {
	bcAPI.conn = wsConn
}

func (bcAPI *WebSocketAPI) Init() {
	bcAPI.ethListener = newEthListener(bcAPI)
}

func (bcAPI *WebSocketAPI) Shutdown() {
	bcAPI.ethListener.Close()
}

func (bcAPI *WebSocketAPI) Name() string {
	return bcAPI.name
}

func (bcAPI *WebSocketAPI) HandleRPC(rpcReq *api.Request) (*api.Response, error) {
	methodName := rpcReq.Method
	resp := &api.Response{}
	if bcAPI.mappings[methodName] == nil {
		fmt.Errorf("Method not supported: %s\n", methodName)
		return nil, errors.New("SRPC Method not supported.")
	}

	// Run the method.
	bcAPI.mappings[methodName](rpcReq, resp)
	// Add a timestamp.
	resp.Timestamp = getTimestamp()
	// The ID is the method being called, for now.
	resp.Id = methodName

	return resp, nil
}

// Add a new method
func (bcAPI *WebSocketAPI) AddMethod(methodName string, method api.WsAPIMethod, replaceOld bool) error {
	if bcAPI.mappings[methodName] != nil {
		if !replaceOld {
			return errors.New("Tried to overwrite an already existing method.")
		} else {
			fmt.Printf("Overwriting old method for '" + methodName + "'.")
		}
	}
	bcAPI.mappings[methodName] = method
	return nil
}

// Remove a method
func (bcAPI *WebSocketAPI) RemoveMethod(methodName string) {
	if bcAPI.mappings[methodName] == nil {
		fmt.Printf("Removal failed. There is no handler for '" + methodName + "'.")
	} else {
		delete(bcAPI.mappings, methodName)
	}
	return
}

func (bcAPI *WebSocketAPI) MyBalance(req *api.Request, resp *api.Response) {
	// TODO add
	retVal := &modules.VString{}
	// TODO Replace with pipe
	myAddr := bcAPI.bc.ActiveAddress()
	retVal.SVal = bcAPI.bc.Account(myAddr).Balance
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) MyAddress(req *api.Request, resp *api.Response) {
	retVal := &modules.VString{}
	// TODO Replace with pipe
	retVal.SVal = bcAPI.bc.ActiveAddress()
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) StartMining(req *api.Request, resp *api.Response) {
	retVal := &modules.VBool{}
	bcAPI.bc.AutoCommit(true)
	retVal.BVal = true
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) StopMining(req *api.Request, resp *api.Response) {
	retVal := &modules.VBool{}
	bcAPI.bc.AutoCommit(false)
	retVal.BVal = true
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) LastBlockNumber(req *api.Request, resp *api.Response) {
	retVal := &modules.VInteger{}
	retVal.IVal = bcAPI.bc.BlockCount() - 1
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) BlockMiniByHash(req *api.Request, resp *api.Response) {
	params := &modules.VString{}
	err := json.Unmarshal(*req.Params, params)

	if err != nil {
		resp.Error = err.Error()
		return
	}

	retVal := &modules.BlockMiniData{}
	fmt.Printf("Block %s\n",params.SVal)
	return
	
	block := bcAPI.bc.Block(params.SVal)
	if block == nil {
		resp.Error = "No block with hash: " + params.SVal
		return
	}

	getBlockMiniDataFromBlock(bcAPI.bc, retVal, block)
	
	resp.Result = retVal
	
}

func (bcAPI *WebSocketAPI) BlockByHash(req *api.Request, resp *api.Response) {
	params := &modules.VString{}
	err := json.Unmarshal(*req.Params, params)

	if err != nil {
		resp.Error = err.Error()
		return
	}

	retVal := &modules.Block{}

	block := bcAPI.bc.Block(params.SVal)
	if block == nil {
		resp.Error = "No block with hash: " + params.SVal
		return
	}
	
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) Account(req *api.Request, resp *api.Response) {
	params := &modules.VString{}
	err := json.Unmarshal(*req.Params, params)

	if err != nil {
		resp.Error = err.Error()
		return
	}

	retVal := bcAPI.bc.Account(params.SVal)
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) Transact(req *api.Request, resp *api.Response) {
	params := &modules.TxIndata{}
	err := json.Unmarshal(*req.Params, params)

	if err != nil {
		resp.Error = err.Error()
		return
	}

	retVal := &modules.TxReceipt{}
	// TODO check sender.
	//err = createTx(bcAPI.ethChain, params.Recipient, params.Value, params.Gas, params.GasCost, params.Data, retVal)
	//if err != nil {
	//	retVal.Error = err.Error()
	//}
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) WorldState(req *api.Request, resp *api.Response) {
	// We do this all in one go.
	blocks := getBlockChain(bcAPI.bc)
	// Let the client know how many blocks there are.
	resp = &api.Response{}
	resp.Id = "NumBlocks"
	resp.Result = &modules.VInteger{IVal: len(blocks) - 1}
	resp.Timestamp = getTimestamp()
	bcAPI.conn.WriteTextMsg(resp)

	// Send blocks one at a time.
	for i := 0; i < len(blocks); i++ {
		resp = &api.Response{}
		resp.Id = "Blocks"
		resp.Result = blocks[i]
		resp.Timestamp = getTimestamp()
		bcAPI.conn.WriteTextMsg(resp)
	}
	
	accounts := bcAPI.bc.WorldState()
	// Let the client know how many accounts there are.
	worldSize := len(accounts.Accounts)
	resp = &api.Response{}
	resp.Id = "NumAccounts"
	resp.Result = &modules.VInteger{IVal: worldSize}
	resp.Timestamp = getTimestamp()
	bcAPI.conn.WriteTextMsg(resp)

	// Send one at a time.
	for _ , hash := range accounts.Order {
		resp = &api.Response{}
		resp.Id = "Accounts"
		acc := accounts.Accounts[hash]
		accMini := &modules.AccountMini{} 
		getAccountMiniFromAccount(accMini,acc)
		resp.Result = accMini
		resp.Timestamp = getTimestamp()
		bcAPI.conn.WriteTextMsg(resp)
	}

	// Finalize.
	resp = &api.Response{}
	resp.Id = "WorldStateDone"
	resp.Result = &modules.NoArgs{}
	resp.Timestamp = getTimestamp()
	bcAPI.conn.WriteTextMsg(resp)
}

type EthListener struct {
	mnk               *WebSocketAPI
	txPreChannel      chan events.Event
	txPreFailChannel  chan events.Event
	txPostChannel     chan events.Event
	txPostFailChannel chan events.Event
	blockChannel      chan events.Event
	stopChannel       chan bool
	logSub            *LogSub
}

func newEthListener(mnk *WebSocketAPI) *EthListener {
	el := &EthListener{}
	el.mnk = mnk
	
	el.blockChannel = make(chan events.Event, 10)
	el.txPreChannel = make(chan events.Event, 10)
	el.txPreFailChannel = make(chan events.Event, 10)
	el.txPostChannel = make(chan events.Event, 10)
	el.txPostFailChannel = make(chan events.Event, 10)
	el.stopChannel = make(chan bool)
	el.blockChannel = el.mnk.bc.Subscribe("","newBlock", "")
	el.txPreChannel = el.mnk.bc.Subscribe("","newTx:pre", "")
	el.txPreFailChannel = el.mnk.bc.Subscribe("","newTx:pre:fail", "")
	el.txPostChannel = el.mnk.bc.Subscribe("","newTx:post", "")
	el.txPostFailChannel = el.mnk.bc.Subscribe("","newTx:post:fail", "")
	
	el.logSub = NewStdLogSub()
	el.logSub.SubId = el.mnk.conn.SessionId()
	el.mnk.ethLogger.AddSub(el.logSub)

	go func(el *EthListener) {
		for {
			select {
			case evt := <-el.blockChannel:
				block, _ := evt.Resource.(*modules.Block)
				fmt.Println("Block added")
				resp := &api.Response{}
				resp.Id = "BlockAdded"
				bd := &modules.BlockMiniData{}
				getBlockMiniDataFromBlock(el.mnk.bc, bd, block)
				resp.Result = bd
				resp.Timestamp = getTimestamp()
				el.mnk.conn.WriteTextMsg(resp)
			case evt := <-el.txPreChannel:
				tx, _ := evt.Resource.(*modules.Transaction)
				resp := &api.Response{}
				resp.Id = "TxPre"
				resp.Result = tx
				resp.Timestamp = getTimestamp()
				el.mnk.conn.WriteTextMsg(resp)
			case evt := <-el.txPreFailChannel:
				tx, _ := evt.Resource.(*modules.Transaction)
				resp := &api.Response{}
				resp.Id = "TxPreFail"
				resp.Result = tx
				resp.Error = tx.Error
				resp.Timestamp = getTimestamp()
				el.mnk.conn.WriteTextMsg(resp)
			case evt := <-el.txPostChannel:
				tx, _ := evt.Resource.(*modules.Transaction)
				resp := &api.Response{}
				resp.Id = "TxPost"
				resp.Result = tx
				resp.Timestamp = getTimestamp()
				el.mnk.conn.WriteTextMsg(resp)
			case evt := <-el.txPostFailChannel:
				tx , _ := evt.Resource.(*modules.Transaction)
				resp := &api.Response{}
				resp.Id = "TxPostFail"
				resp.Result = tx
				resp.Error = tx.Error
				resp.Timestamp = getTimestamp()
				el.mnk.conn.WriteTextMsg(resp)
			case txt := <-el.logSub.Channel:
				resp := &api.Response{}
				resp.Id = "Log"
				resp.Result = &modules.VString{SVal: txt}
				resp.Timestamp = getTimestamp()
				el.mnk.conn.WriteTextMsg(resp)
			case <-el.stopChannel:
				// Quit this
				return
			}
		}
	}(el)
	return el
}

func (el *EthListener) Close() {
	//el.mnk.bc.Unsubscribe()
	/*
	rctr := el.mnk.ethChain.Ethereum.Reactor()
	rctr.Unsubscribe("newBlock", el.blockChannel)
	rctr.Unsubscribe("newTx:pre", el.txPreChannel)
	rctr.Unsubscribe("newTx:pre:fail", el.txPreFailChannel)
	rctr.Unsubscribe("newTx:post", el.txPostChannel)
	rctr.Unsubscribe("newTx:post:fail", el.txPostFailChannel)
	el.mnk.ethLogger.RemoveSub(el.logSub)
	*/
}

func getTimestamp() int {
	return int(time.Now().In(time.UTC).UnixNano() >> 6)
}

// TODO while testing
type LogSub struct {
	Channel  chan string
	SubId    uint32
	LogLevel monklog.LogLevel
	Enabled  bool
}

func NewStdLogSub() *LogSub {
	ls := &LogSub{
		Channel:  make(chan string),
		SubId:    0,
		LogLevel: monklog.LogLevel(5),
		Enabled:  true,
	}
	return ls
}

type EthLogger struct {
	mutex     *sync.Mutex
	logReader io.Reader
	logWriter io.Writer
	logLevel  monklog.LogLevel
	subs      []*LogSub
}

func NewEthLogger() *EthLogger {
	el := &EthLogger{}
	el.mutex = &sync.Mutex{};
	el.logLevel = monklog.LogLevel(5)
	el.logReader, el.logWriter = io.Pipe()

	monklog.AddLogSystem(monklog.NewStdLogSystem(el.logWriter, log.LstdFlags, el.logLevel))

	go func(el *EthLogger) {
		scanner := bufio.NewScanner(el.logReader)
		for scanner.Scan() {
			text := scanner.Text()
			el.mutex.Lock()
			for _, sub := range el.subs {
				sub.Channel <- text
			}
			el.mutex.Unlock()
		}
	}(el)
	return el
}

func (el *EthLogger) AddSub(sub *LogSub) {
	el.mutex.Lock()
	el.subs = append(el.subs, sub)
	el.mutex.Unlock()
}

func (el *EthLogger) RemoveSub(sub *LogSub) {
	el.mutex.Lock()
	theIdx := -1
	for idx, s := range el.subs {
		if sub.SubId == s.SubId {
			theIdx = idx
			break
		}
	}
	if theIdx >= 0 {
		el.subs = append(el.subs[:theIdx], el.subs[theIdx+1:]...)
	}
	el.mutex.Unlock()
}
