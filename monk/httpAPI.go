package monk

import (
	"encoding/hex"
	"fmt"
	"github.com/eris-ltd/thelonious/ethutil"
	"github.com/eris-ltd/thelonious/monk"
	"net/http"
)

type Monk struct {
	EthChain *monk.EthChain
}

func (mapi *Monk) IsContract(r *http.Request, args *VString, reply *VBool) error {
	reply.BVal = isContract(mapi.EthChain, args.SVal)
	return nil
}

func (mapi *Monk) BalanceAt(r *http.Request, args *VString, reply *VString) error {
	sHex, err := hex.DecodeString(args.SVal)
	if err != nil {
		fmt.Println(err.Error())
		reply.SVal = ERR_MALFORMED_ADDRESS
		return nil
	}
	balance := mapi.EthChain.Pipe.Balance(sHex)
	reply.SVal = balance.String()
	return nil
}

func (mapi *Monk) MyBalance(r *http.Request, args *NoArgs, reply *VString) error {
	myAddr := mapi.EthChain.Ethereum.KeyManager().Address()
	balance := mapi.EthChain.Pipe.Balance(myAddr)
	reply.SVal = balance.String()
	return nil
}

func (mapi *Monk) StorageAt(r *http.Request, args *StateAtArgs, reply *VString) error {
	stateobj := getStateObject(mapi.EthChain, args.Address)
	if stateobj == nil {
		reply.SVal = ERR_NO_SUCH_ADDRESS
		return nil
	}
	storage := stateobj.GetStorage(ethutil.Big(args.Storage))
	if storage == nil {
		reply.SVal = ERR_STATE_NO_STORAGE
		return nil
	}
	reply.SVal = storage.String()
	return nil
}

// TODO min gascost is hardcoded in block_chain NewBlock. Will this vary at some point?
func (mapi *Monk) MinGascost(r *http.Request, args *NoArgs, reply *VString) error {
	reply.SVal = "10000000000000"
	return nil
}

func (mapi *Monk) StartMining(r *http.Request, args *NoArgs, reply *VBool) error {
	reply.BVal = mapi.EthChain.StartMining()
	return nil
}

func (mapi *Monk) StopMining(r *http.Request, args *NoArgs, reply *VBool) error {
	reply.BVal = mapi.EthChain.StopMining()
	return nil
}

func (mapi *Monk) IsMining(r *http.Request, args *NoArgs, reply *VBool) error {
	reply.BVal = mapi.EthChain.Ethereum.IsMining()
	return nil
}

func (mapi *Monk) MyAddress(r *http.Request, args *NoArgs, reply *VString) error {
	reply.SVal = hex.EncodeToString(mapi.EthChain.Ethereum.KeyManager().Address())
	return nil
}

func (mapi *Monk) BlockLatest(r *http.Request, args *NoArgs, reply *BlockData) error {
	addr, _ := hex.DecodeString("29c8e2e2a699ed64296025795b5dca20647c66de")
	acc := mapi.EthChain.Ethereum.StateManager().CurrentState().GetAccount(addr)
	if acc == nil {
		fmt.Println("No such account.")
	}
	fmt.Printf("%x\n", acc.CodeHash)
	lbh := mapi.EthChain.Ethereum.BlockChain().LastBlockHash
	argz := &VString{SVal: hex.EncodeToString(lbh)}
	mapi.BlockByHash(r, argz, reply)
	return nil
}

func (mapi *Monk) BlockByHash(r *http.Request, args *VString, reply *BlockData) error {
	// Get the block.
	bts, err := hex.DecodeString(args.SVal)
	if err != nil {
		fmt.Println(err.Error())
		reply.Hash = ERR_MALFORMED_TX_HASH
		return nil
	}

	block := mapi.EthChain.Ethereum.BlockChain().GetBlock(bts)

	if block == nil {
		reply.Hash = ERR_NO_SUCH_BLOCK
		return nil
	}
	getBlockDataFromBlock(reply, block)
	return nil
}

func (mapi *Monk) Transact(r *http.Request, args *TxIndata, reply *TxReceipt) error {
	err := createTx(mapi.EthChain, args.Recipient, args.Value, args.Gas, args.GasCost, args.Data, reply)

	if err != nil {
		reply.Error = err.Error()
	}
	return nil
}

func (mapi *Monk) Account(r *http.Request, args *VString, reply *Account) error {

	// Get the block.
	addr, err := hex.DecodeString(args.SVal)
	if err != nil {
		fmt.Println(err.Error())
		reply.Address = ERR_MALFORMED_ADDRESS
		return nil
	}

	so := mapi.EthChain.Ethereum.BlockChain().CurrentBlock.State().GetOrNewStateObject(addr)
	getAccountFromStateObject(reply, so)
	return nil
}
