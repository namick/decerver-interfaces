package monk

// This handles socket-based rpc. Part of it is reacting to requests sent from the
// client, and part of it is reacting to changes in the ethereum world state,
// and propagating these.
import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eris-ltd/deCerver-interfaces/api"
	"github.com/eris-ltd/thelonious/ethchain"
	"github.com/eris-ltd/thelonious/ethreact"
	"github.com/eris-ltd/thelonious/monk"
	"github.com/golang/glog"
	"time"
)

type MonkWsAPIFactory struct {
	ethChain    *monk.EthChain
	serviceName string
}

func NewMonkWsAPIFactory(ethChain *monk.EthChain) *MonkWsAPIFactory {
	fact := &MonkWsAPIFactory{
		ethChain:    ethChain,
		serviceName: "MonkWsAPI",
	}
	return fact
}

func (fact *MonkWsAPIFactory) Init() {

}

func (fact *MonkWsAPIFactory) Shutdown() {
	// TODO fix
	//fact.ethChain.Stop()
}

func (fact *MonkWsAPIFactory) ServiceName() string {
	return fact.serviceName
}

func (fact *MonkWsAPIFactory) CreateService() api.WsAPIService {
	ec := monk.NewEth(fact.ethChain.Ethereum)
	ec.Init()
	service := newMonkWsAPI(ec)
	service.name = fact.serviceName
	return service
}

type MonkWsAPI struct {
	name        string
	mappings    map[string]api.WsAPIMethod
	ethChain    *monk.EthChain
	conn        api.WebSocketObj
	ethListener *EthListener
}

// Create a new handler
func newMonkWsAPI(eth *monk.EthChain) *MonkWsAPI {
	esrpc := &MonkWsAPI{}
	esrpc.ethChain = eth

	esrpc.mappings = make(map[string]api.WsAPIMethod)
	esrpc.mappings["MyBalance"] = esrpc.MyBalance
	esrpc.mappings["MyAddress"] = esrpc.MyAddress
	esrpc.mappings["StartMining"] = esrpc.StartMining
	esrpc.mappings["StopMining"] = esrpc.StopMining
	esrpc.mappings["LastBlockNumber"] = esrpc.LastBlockNumber
	esrpc.mappings["BlockByHash"] = esrpc.BlockByHash
	esrpc.mappings["Account"] = esrpc.Account
	esrpc.mappings["Transact"] = esrpc.Transact
	esrpc.mappings["WorldState"] = esrpc.WorldState

	return esrpc
}

func (esrpc *MonkWsAPI) SetConnection(wsConn api.WebSocketObj) {
	esrpc.conn = wsConn
}

func (esrpc *MonkWsAPI) Init() {
	esrpc.ethListener = newEthListener(esrpc)
}

func (esrpc *MonkWsAPI) Shutdown() {
	esrpc.ethListener.Close()
}

func (esrpc *MonkWsAPI) Name() string {
	return esrpc.name
}

func (esrpc *MonkWsAPI) HandleRPC(rpcReq *api.Request) (*api.Response, error) {
	methodName := rpcReq.Method
	resp := &api.Response{}
	if esrpc.mappings[methodName] == nil {
		fmt.Errorf("Method not supported: %s\n", methodName)
		return nil, errors.New("SRPC Method not supported.")
	}

	// Run the method.
	esrpc.mappings[methodName](rpcReq, resp)
	// Add a timestamp.
	resp.Timestamp = getTimestamp()
	// The ID is the method being called, for now.
	resp.Id = methodName

	return resp, nil
}

// Add a new method
func (esrpc *MonkWsAPI) AddMethod(methodName string, method api.WsAPIMethod, replaceOld bool) error {
	if esrpc.mappings[methodName] != nil {
		if !replaceOld {
			return errors.New("Tried to overwrite an already existing method.")
		} else {
			glog.Infoln("Overwriting old method for '" + methodName + "'.")
		}

	}
	esrpc.mappings[methodName] = method
	return nil
}

// Remove a method
func (esrpc *MonkWsAPI) RemoveMethod(methodName string) {
	if esrpc.mappings[methodName] == nil {
		glog.Infoln("Removal failed. There is no handler for '" + methodName + "'.")
	} else {
		delete(esrpc.mappings, methodName)
	}
	return
}

func (esrpc *MonkWsAPI) MyBalance(req *api.Request, resp *api.Response) {
	retVal := &VString{}
	// TODO Replace with pipe
	myAddr := esrpc.ethChain.Ethereum.KeyManager().Address()
	balance := esrpc.ethChain.Pipe.Balance(myAddr)
	// -----------------
	retVal.SVal = balance.String()
	resp.Result = retVal
}

func (esrpc *MonkWsAPI) MyAddress(req *api.Request, resp *api.Response) {
	retVal := &VString{}
	retVal.SVal = hex.EncodeToString(esrpc.ethChain.Ethereum.KeyManager().Address())
	resp.Result = retVal
}

func (esrpc *MonkWsAPI) StartMining(req *api.Request, resp *api.Response) {
	retVal := &VBool{}
	retVal.BVal = esrpc.ethChain.StartMining()
	resp.Result = retVal
}

func (esrpc *MonkWsAPI) StopMining(req *api.Request, resp *api.Response) {
	retVal := &VBool{}
	retVal.BVal = esrpc.ethChain.StopMining()
	resp.Result = retVal
}

func (esrpc *MonkWsAPI) LastBlockNumber(req *api.Request, resp *api.Response) {
	retVal := &VInteger{}
	retVal.IVal = getLastBlockNumber(esrpc.ethChain)
	resp.Result = retVal
}

func (esrpc *MonkWsAPI) BlockByHash(req *api.Request, resp *api.Response) {
	params := &VString{}
	err := json.Unmarshal(*req.Params, params)

	if err != nil {
		resp.Error = err.Error()
		return
	}

	retVal := &BlockData{}
	hash, decErr := hex.DecodeString(params.SVal)

	if decErr != nil {
		resp.Error = decErr.Error()
		return
	}

	block := esrpc.ethChain.Pipe.Block(hash)
	if block == nil {
		resp.Error = "No block with hash: " + params.SVal
		return
	}

	getBlockDataFromBlock(retVal, block)
	resp.Result = retVal
}

func (esrpc *MonkWsAPI) Account(req *api.Request, resp *api.Response) {
	params := &VString{}
	err := json.Unmarshal(*req.Params, params)

	if err != nil {
		resp.Error = err.Error()
		return
	}

	retVal := &Account{}
	addr, decErr := hex.DecodeString(params.SVal)

	if decErr != nil {
		resp.Error = decErr.Error()
		return
	}

	curBlock := esrpc.ethChain.Ethereum.BlockChain().CurrentBlock
	account := curBlock.State().GetStateObject(addr)
	if account == nil {
		resp.Error = "No account with address: " + params.SVal
		return
	}

	getAccountFromStateObject(retVal, account)
	resp.Result = retVal
}

/*
type TxIndata struct {
	Recipient string
	Value     string
	Gas       string
	GasCost   string
	Data      string
}
*/

func (esrpc *MonkWsAPI) Transact(req *api.Request, resp *api.Response) {
	params := &TxIndata{}
	err := json.Unmarshal(*req.Params, params)

	if err != nil {
		resp.Error = err.Error()
		return
	}

	retVal := &TxReceipt{}
	// TODO check sender.
	err = createTx(esrpc.ethChain, params.Recipient, params.Value, params.Gas, params.GasCost, params.Data, retVal)
	if err != nil {
		retVal.Error = err.Error()
	}
	resp.Result = retVal
}

func (esrpc *MonkWsAPI) WorldState(req *api.Request, resp *api.Response) {

	blocks := getWorldState(esrpc.ethChain)
	// Let the client know how many blocks there are.
	resp = &api.Response{}
	resp.Id = "NumBlocks"
	resp.Result = &VInteger{IVal: len(blocks) - 1}
	resp.Timestamp = getTimestamp()
	esrpc.conn.WriteTextMsg(resp)

	// Send blocks one at a time.
	for i := 0; i < len(blocks); i++ {
		resp = &api.Response{}
		resp.Id = "Blocks"
		resp.Result = blocks[i]
		resp.Timestamp = getTimestamp()
		esrpc.conn.WriteTextMsg(resp)
	}

	accounts := getAccounts(esrpc.ethChain)

	// Let the client know how many accounts there are.
	resp = &api.Response{}
	resp.Id = "NumAccounts"
	resp.Result = &VInteger{IVal: len(accounts)}
	resp.Timestamp = getTimestamp()
	esrpc.conn.WriteTextMsg(resp)

	// Dispatch these one at a time, and also register listeners to all these addresses.
	for i := 0; i < len(accounts); i++ {
		resp = &api.Response{}
		resp.Id = "Accounts"
		resp.Result = accounts[i]
		resp.Timestamp = getTimestamp()
		esrpc.conn.WriteTextMsg(resp)
	}

	// Finalize.
	resp = &api.Response{}
	resp.Id = "WorldStateDone"
	resp.Result = &NoArgs{}
	resp.Timestamp = getTimestamp()
	esrpc.conn.WriteTextMsg(resp)
}

type EthListener struct {
	mnk               *MonkWsAPI
	txPreChannel      chan ethreact.Event
	txPreFailChannel  chan ethreact.Event
	txPostChannel     chan ethreact.Event
	txPostFailChannel chan ethreact.Event
	blockChannel      chan ethreact.Event
	stopChannel       chan bool
}

func newEthListener(mnk *MonkWsAPI) *EthListener {
	el := &EthListener{}
	el.mnk = mnk

	el.blockChannel = make(chan ethreact.Event, 10)
	el.txPreChannel = make(chan ethreact.Event, 10)
	el.txPreFailChannel = make(chan ethreact.Event, 10)
	el.txPostChannel = make(chan ethreact.Event, 10)
	el.txPostFailChannel = make(chan ethreact.Event, 10)
	el.stopChannel = make(chan bool)
	el.mnk.ethChain.Ethereum.Reactor().Subscribe("newBlock", el.blockChannel)
	el.mnk.ethChain.Ethereum.Reactor().Subscribe("newTx:pre", el.txPreChannel)
	el.mnk.ethChain.Ethereum.Reactor().Subscribe("newTx:pre:fail", el.txPreFailChannel)
	el.mnk.ethChain.Ethereum.Reactor().Subscribe("newTx:post", el.txPostChannel)
	el.mnk.ethChain.Ethereum.Reactor().Subscribe("newTx:post:fail", el.txPostFailChannel)

	go func(el *EthListener) {
		for {
			select {
			case evt := <-el.blockChannel:
				block, _ := evt.Resource.(*ethchain.Block)
				fmt.Println("Block added")
				resp := &api.Response{}
				resp.Id = "BlockAdded"
				bd := &BlockMiniData{}
				getBlockMiniDataFromBlock(el.mnk.ethChain, bd, block)
				resp.Result = bd
				resp.Timestamp = getTimestamp()
				el.mnk.conn.WriteTextMsg(resp)
			case evt := <-el.txPreChannel:
				tx, _ := evt.Resource.(*ethchain.Transaction)
				resp := &api.Response{}
				resp.Id = "TxPre"
				trans := &Transaction{}
				getTransactionFromTx(trans, tx)
				resp.Result = trans
				resp.Timestamp = getTimestamp()
				el.mnk.conn.WriteTextMsg(resp)
			case evt := <-el.txPreFailChannel:
				txFail, _ := evt.Resource.(*ethchain.TxFail)
				resp := &api.Response{}
				resp.Id = "TxPreFail"
				trans := &Transaction{}
				getTransactionFromTx(trans, txFail.Tx)
				trans.Error = txFail.Err.Error()
				resp.Result = trans
				resp.Timestamp = getTimestamp()
				el.mnk.conn.WriteTextMsg(resp)
			case evt := <-el.txPostChannel:
				tx, _ := evt.Resource.(*ethchain.Transaction)
				resp := &api.Response{}
				resp.Id = "TxPost"
				trans := &Transaction{}
				getTransactionFromTx(trans, tx)
				resp.Result = trans
				resp.Timestamp = getTimestamp()
				el.mnk.conn.WriteTextMsg(resp)
			case evt := <-el.txPostFailChannel:
				txFail, _ := evt.Resource.(*ethchain.TxFail)
				resp := &api.Response{}
				resp.Id = "TxPostFail"
				trans := &Transaction{}
				getTransactionFromTx(trans, txFail.Tx)
				trans.Error = txFail.Err.Error()
				resp.Result = trans
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
	rctr := el.mnk.ethChain.Ethereum.Reactor()
	rctr.Unsubscribe("newBlock", el.blockChannel)
	rctr.Unsubscribe("newTx:pre", el.txPreChannel)
	rctr.Unsubscribe("newTx:pre:fail", el.txPreFailChannel)
	rctr.Unsubscribe("newTx:post", el.txPostChannel)
	rctr.Unsubscribe("newTx:post:fail", el.txPostFailChannel)
}

func getTimestamp() int {
	return int(time.Now().In(time.UTC).UnixNano() >> 6)
}
