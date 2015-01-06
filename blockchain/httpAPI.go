package blockchain

/*
import (
	"github.com/eris-ltd/decerver-interfaces/modules"
	"net/http"
)

type Monk struct {
	bc modules.Blockchain
}

func (mapi *Monk) IsContract(r *http.Request, args *modules.VString, reply *modules.VBool) error {
	acc := mapi.bc.Account(args.SVal)
	if acc == nil {
		reply.BVal = false
	} else {
		reply.BVal = acc["IsScript"].(bool)
	}
	return nil
}

func (mapi *Monk) BalanceAt(r *http.Request, args *modules.VString, reply *modules.VString) error {

	acc := mapi.bc.Account(args.SVal)
	if acc == nil {
		reply.SVal = ""
	} else {
		reply.SVal = acc["Balance"].(string)
	}

	return nil
}

func (mapi *Monk) MyBalance(r *http.Request, args *modules.NoArgs, reply *modules.VString) error {
	// TODO add
	return nil
}

func (mapi *Monk) StorageAt(r *http.Request, args *modules.StateAtArgs, reply *modules.VString) error {
	reply.SVal = mapi.bc.StorageAt(args.Address,args.Storage)
	return nil
}

// TODO min gascost is hardcoded in block_chain NewBlock. Will this vary at some point?
func (mapi *Monk) MinGascost(r *http.Request, args *modules.NoArgs, reply *modules.VString) error {
	reply.SVal = "10000000000000"
	return nil
}

func (mapi *Monk) StartMining(r *http.Request, args *modules.NoArgs, reply *modules.VBool) error {
	mapi.bc.AutoCommit(true)
	return nil
}

func (mapi *Monk) StopMining(r *http.Request, args *modules.NoArgs, reply *modules.VBool) error {
	mapi.bc.AutoCommit(false)
	return nil
}

func (mapi *Monk) IsMining(r *http.Request, args *modules.NoArgs, reply *modules.VBool) error {
	reply.BVal = mapi.bc.IsAutocommit()
	return nil
}

func (mapi *Monk) MyAddress(r *http.Request, args *modules.NoArgs, reply *modules.VString) error {
	// TODO add
	return nil
}

func (mapi *Monk) BlockLatest(r *http.Request, args *modules.NoArgs, reply *modules.Block) error {
	// TODO add
	return nil
}

func (mapi *Monk) BlockByHash(r *http.Request, args *modules.VString, reply *modules.Block) error {
	mapi.bc.Block(args.SVal)
	return nil
}

func (mapi *Monk) Transact(r *http.Request, args *modules.TxIndata, reply *modules.TxReceipt) error {
	// TODO add
	//err := createTx(mapi.MonkModule, args.Recipient, args.Value, args.Gas, args.GasCost, args.Data, reply)

	//if err != nil {
	//	reply.Error = err.Error()
	//}
	return nil
}

func (mapi *Monk) Account(r *http.Request, args *modules.VString, reply *modules.Account) error {
 	// TODO add
	return nil
}
*/
