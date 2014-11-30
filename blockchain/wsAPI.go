package blockchain
/*

// This handles socket-based rpc. Part of it is reacting to requests sent from the
// client, and part of it is reacting to changes in the ethereum world state,
// and propagating these.
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eris-ltd/decerver-interfaces/api"
	"github.com/eris-ltd/decerver-interfaces/events"
	"github.com/eris-ltd/decerver-interfaces/modules"
	"github.com/eris-ltd/decerver-interfaces/util"
	"strconv"
	"strings"
	"time"
)

type WebSocketAPIFactory struct {
	bc          modules.Blockchain
	serviceName string
}

func NewWebSocketAPIFactory(bc modules.Blockchain) *WebSocketAPIFactory {
	fact := &WebSocketAPIFactory{
		bc:          bc,
		serviceName: "BlockchainWs",
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
	return service
}

type WebSocketAPI struct {
	name       string
	mappings   map[string]api.WsAPIMethod
	bc         modules.Blockchain
	session    api.WsSession
	bcListener *BcListener
	blockQueue *util.BlockMiniQueue
	wsUpdated  bool
}

// Create a new handler
func newWebSocketAPI(bc modules.Blockchain) *WebSocketAPI {
	
	bcAPI := &WebSocketAPI{}
	bcAPI.bc = bc
	bcAPI.blockQueue = util.NewBlockMiniQueue()
	bcAPI.wsUpdated = false

	bcAPI.mappings = make(map[string]api.WsAPIMethod)
	bcAPI.mappings["MyBalance"] = bcAPI.MyBalance
	bcAPI.mappings["ActiveAddress"] = bcAPI.ActiveAddress
	bcAPI.mappings["MyAddresses"] = bcAPI.MyAddresses
	bcAPI.mappings["NewAddress"] = bcAPI.NewAddress
	bcAPI.mappings["SetAddress"] = bcAPI.SetAddress
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

func (bcAPI *WebSocketAPI) SetSession(wsSession api.WsSession) {
	bcAPI.session = wsSession
}

func (bcAPI *WebSocketAPI) Init() {
	bcAPI.bcListener = newBcListener(bcAPI)
}

func (bcAPI *WebSocketAPI) Shutdown() {
	bcAPI.bcListener.Close()
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

func (bcAPI *WebSocketAPI) ActiveAddress(req *api.Request, resp *api.Response) {
	retVal := &modules.VString{}
	// TODO Replace with pipe
	retVal.SVal = bcAPI.bc.ActiveAddress()
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) MyAddresses(req *api.Request, resp *api.Response) {
	retVal := &modules.Addresses{}
	retVal.ActiveAddress = bcAPI.bc.ActiveAddress()
	numKeys := bcAPI.bc.AddressCount()
	keyArr := make([]string, numKeys)
	var err error
	for idx := 0; idx < numKeys; idx++ {
		keyArr[idx] , err = bcAPI.bc.Address(idx)
		if err != nil {
			resp.Error = err.Error()
			break
		}
	}
	retVal.AddressList = keyArr
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) NewAddress(req *api.Request, resp *api.Response) {
	retVal := &modules.VString{}
	// TODO Replace with pipe
	retVal.SVal = bcAPI.bc.NewAddress(true)
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) SetAddress(req *api.Request, resp *api.Response) {
	
	params := &modules.VString{}
	err := json.Unmarshal(*req.Params, params)

	if err != nil {
		resp.Error = err.Error()
		return
	}
	
	retVal := &modules.VString{}
	// TODO Replace with pipe
	err = bcAPI.bc.SetAddress(params.SVal)
	
	if err != nil {
		retVal.SVal = err.Error()
	}
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

	retVal := &modules.BlockMini{}
	fmt.Printf("Block %s\n", params.SVal)
	return

	block := bcAPI.bc.Block(params.SVal)
	if block == nil {
		resp.Error = "No block with hash: " + params.SVal
		return
	}

	getBlockMiniFromBlock(bcAPI.bc, retVal, block)

	resp.Result = retVal

}

func (bcAPI *WebSocketAPI) BlockByHash(req *api.Request, resp *api.Response) {
	params := &modules.VString{}
	err := json.Unmarshal(*req.Params, params)

	if err != nil {
		resp.Error = err.Error()
		return
	}
	//params.SVal = "0x" + params.SVal
	fmt.Println("Block being fetched: " + params.SVal)

	block := bcAPI.bc.Block(params.SVal)
	if block == nil {
		resp.Error = "No block with hash: " + params.SVal
		return
	}

	resp.Result = block
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
		fmt.Printf("Tx indata error: %s\n", err.Error())
		resp.Error = err.Error()
		return
	}

	fmt.Printf("Tx indata: %v\n", params)

	retVal := &modules.TxReceipt{}

	// Contract create
	if params.Recipient == "" {
		fmt.Println("Processing contract create tx")
		addr, err := bcAPI.bc.Script(params.Data, "lll-literal")
		if err != nil {
			retVal.Compiled = false
			retVal.Error = err.Error()
			retVal.Success = false
		} else {
			retVal.Address = addr
			retVal.Compiled = true
			retVal.Success = true
		}
		// Tx
	} else if params.Data == "" {
		fmt.Println("Processing tx")
		hash, _ := bcAPI.bc.Tx(params.Recipient, params.Value)
		retVal.Success = true
		retVal.Hash = hash
		// It's a message
	} else if params.Value == "" {
		fmt.Println("Processing message")
		txData := strings.Split(params.Data, "\n")
		for idx, val := range txData {
			txData[idx] = strings.Trim(val, " ")
		}

		hash, _ := bcAPI.bc.Msg(params.Recipient, txData)
		retVal.Success = true
		retVal.Hash = hash
	} else {
		// TODO general purpose transaction.
		fmt.Println("Processing message")
		txData := strings.Split(params.Data, "\n")
		for idx, val := range txData {
			txData[idx] = strings.Trim(val, " ")
		}
		hash, _ := bcAPI.bc.Msg(params.Recipient, txData)
		retVal.Success = true
		retVal.Hash = hash
	}
	resp.Result = retVal
}

func (bcAPI *WebSocketAPI) WorldState(req *api.Request, resp *api.Response) {
	// We do this all in one go.
	blocks := getBlockChain(bcAPI.bc)
	// Let the client know how many blocks there are.
	resp = &api.Response{}
	resp.Id = "NumBlocks"
	resp.Result = &modules.VInteger{IVal: len(blocks) - 1}
	bcAPI.session.WriteJsonMsg(resp)

	// Send blocks one at a time.
	for i := 0; i < len(blocks); i++ {
		resp = &api.Response{}
		resp.Id = "Blocks"
		resp.Result = blocks[i]
		bcAPI.session.WriteJsonMsg(resp)
		time.Sleep(50)
	}

	accounts := bcAPI.bc.WorldState()
	// Let the client know how many accounts there are.
	worldSize := len(accounts.Accounts)
	resp = &api.Response{}
	resp.Id = "NumAccounts"
	resp.Result = &modules.VInteger{IVal: worldSize}
	bcAPI.session.WriteJsonMsg(resp)

	// Send one at a time.
	for _, hash := range accounts.Order {
		resp = &api.Response{}
		resp.Id = "Accounts"
		acc := accounts.Accounts[hash]
		accMini := &modules.AccountMini{}
		getAccountMiniFromAccount(accMini, acc)
		resp.Result = accMini
		bcAPI.session.WriteJsonMsg(resp)
		time.Sleep(50)
	}

	time.Sleep(200)

	// Now flush the generated block queue
	for !bcAPI.blockQueue.IsEmpty() {
		// Finalize.
		resp = &api.Response{}
		resp.Id = "BlockAdded"
		resp.Result = bcAPI.blockQueue.Pop()
		bcAPI.session.WriteJsonMsg(resp)
	}

	bcAPI.wsUpdated = true

	// Finalize.
	resp = &api.Response{}
	resp.Id = "WorldStateDone"
	resp.Result = &modules.NoArgs{}
	bcAPI.session.WriteJsonMsg(resp)

}

// This object is used to subscribe directly to the blockchain rather then going through
// the global eventprocessor.
type BcListener struct {
	bcAPI             *WebSocketAPI
	txPreChannel      chan events.Event
	txPreFailChannel  chan events.Event
	txPostChannel     chan events.Event
	txPostFailChannel chan events.Event
	blockChannel      chan events.Event
	stopChannel       chan bool
}

func newBcListener(bcAPI *WebSocketAPI) *BcListener {
	bl := &BcListener{}
	bl.bcAPI = bcAPI

	bl.blockChannel = make(chan events.Event, 10)
	bl.txPreChannel = make(chan events.Event, 10)
	bl.txPreFailChannel = make(chan events.Event, 10)
	bl.txPostChannel = make(chan events.Event, 10)
	bl.txPostFailChannel = make(chan events.Event, 10)
	bl.stopChannel = make(chan bool)
	idStr := strconv.Itoa(int(bl.bcAPI.session.SessionId()))
	c := "newBlock"
	bl.blockChannel = bl.bcAPI.bc.Subscribe(c+idStr, c, "")
	c = "newTx:pre"
	bl.txPreChannel = bl.bcAPI.bc.Subscribe(c+idStr, c, "")
	c = "newTx:pre:fail"
	bl.txPreFailChannel = bl.bcAPI.bc.Subscribe(c+idStr, c, "")
	c = "newTx:post"
	bl.txPostChannel = bl.bcAPI.bc.Subscribe(c+idStr, c, "")
	c = "newTx:post:fail"
	bl.txPostFailChannel = bl.bcAPI.bc.Subscribe(c+idStr, c, "")

	go func(bl *BcListener) {
		for {
			select {
			case evt := <-bl.blockChannel:
				block, _ := evt.Resource.(*modules.Block)
				if block == nil {
					continue;
				} 
				fmt.Println("Block added")
				resp := &api.Response{}
				resp.Id = "BlockAdded"
				bd := &modules.BlockMini{}
				getBlockMiniFromBlock(bl.bcAPI.bc, bd, block)
				if bl.bcAPI.wsUpdated == false {
					bl.bcAPI.blockQueue.Push(bd)
				} else {
					resp.Result = bd
					bl.bcAPI.session.WriteJsonMsg(resp)
				}
			case evt := <-bl.txPreChannel:
				tx, _ := evt.Resource.(*modules.Transaction)
				if tx == nil {
					continue;
				}
				resp := &api.Response{}
				resp.Id = "TxPre"
				resp.Result = tx
				bl.bcAPI.session.WriteJsonMsg(resp)
			case evt := <-bl.txPreFailChannel:
				tx, _ := evt.Resource.(*modules.Transaction)
				if tx == nil {
					continue;
				}
				resp := &api.Response{}
				resp.Id = "TxPreFail"
				resp.Result = tx
				resp.Error = tx.Error
				bl.bcAPI.session.WriteJsonMsg(resp)
			case evt := <-bl.txPostChannel:
				tx, _ := evt.Resource.(*modules.Transaction)
				if tx == nil {
					continue;
				}
				resp := &api.Response{}
				resp.Id = "TxPost"
				resp.Result = tx
				bl.bcAPI.session.WriteJsonMsg(resp)
			case evt := <-bl.txPostFailChannel:
				tx, _ := evt.Resource.(*modules.Transaction)
				if tx == nil {
					continue;
				}
				resp := &api.Response{}
				resp.Id = "TxPostFail"
				resp.Result = tx
				resp.Error = tx.Error
				bl.bcAPI.session.WriteJsonMsg(resp)
			case <-bl.stopChannel:
				// Quit this
				return
			}
		}
	}(bl)
	return bl
}

func (bl *BcListener) Close() {
	idStr := strconv.Itoa(int(bl.bcAPI.session.SessionId()))
	c := "newBlock"
	fmt.Printf("Unregister: " + c + idStr)
	bl.bcAPI.bc.UnSubscribe(c + idStr)
	c = "newTx:pre"
	fmt.Printf("Unregister: " + c + idStr)
	bl.bcAPI.bc.UnSubscribe(c + idStr)
	c = "newTx:pre:fail"
	fmt.Printf("Unregister: " + c + idStr)
	bl.bcAPI.bc.UnSubscribe(c + idStr)
	c = "newTx:post"
	fmt.Printf("Unregister: " + c + idStr)
	bl.bcAPI.bc.UnSubscribe(c + idStr)
	c = "newTx:post:fail"
	fmt.Printf("Unregister: " + c + idStr)
	bl.bcAPI.bc.UnSubscribe(c + idStr)
}
*/