package monkrpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	_ "strconv"

	"github.com/eris-ltd/thelonious/monkchain"
	"github.com/eris-ltd/thelonious/monkcrypto"
	"github.com/eris-ltd/thelonious/monkdb"
	"github.com/eris-ltd/thelonious/monklog"
	"github.com/eris-ltd/thelonious/monkrpc"
	"github.com/eris-ltd/thelonious/monkutil"
)

var (
	GoPath = os.Getenv("GOPATH")
	usr, _ = user.Current() // error?!
)

// A tx to be signed by a local daemon
func (mod *MonkRpcModule) newLocalTx(addr, value, gas, gasprice, body string) monkrpc.NewTxArgs {
	return monkrpc.NewTxArgs{
		Recipient: addr,
		Value:     value,
		Gas:       gas,
		GasPrice:  gasprice,
		Body:      body,
	}
}

// A full formed and signed rlp encoded tx to be broadcast by a remote server
func (mod *MonkRpcModule) newRemoteTx(keys *monkcrypto.KeyPair, addr, value, gas, gasprice, body string) monkrpc.PushTxArgs {
	addrB := monkutil.Hex2Bytes(addr)
	valB := monkutil.Big(value)
	gasB := monkutil.Big(gas)
	gaspriceB := monkutil.Big(gasprice)
	bodyB := monkutil.Hex2Bytes(body)

	// get nonce
	args := monkrpc.GetTxCountArgs{monkutil.Bytes2Hex(keys.Address())}
	n, _ := mod.rpcTxCountCall(args)
	fmt.Println(n)

	tx := monkchain.NewTransactionMessage(addrB, valB, gasB, gaspriceB, bodyB)
	tx.Nonce = n
	tx.Sign(keys.PrivateKey)
	txenc := tx.RlpEncode()
	return monkrpc.PushTxArgs{monkutil.Bytes2Hex(txenc)}
}

// TODO: This is awful, just awful, terribly, terribly awful
func (mod *MonkRpcModule) rpcTxCountCall(args monkrpc.GetTxCountArgs) (uint64, error) {
	res := new(string)
	err := mod.client.Call("TheloniousApi.GetTxCountAt", args, res)
	if err != nil {
		return 0, err
	}
	fmt.Println(*res)
	r := new(monkrpc.SuccessRes)
	err = json.Unmarshal([]byte(*res), r)
	if err != nil {
		log.Fatal(err)
	}
	resMap := r.Result.(map[string]interface{})
	n := resMap["nonce"].(float64) // WTF?!?!?!?

	// ok, this was an abomination of a clean rpc call
	// but hey, fuck you, it works, please make it cleaner if you know how
	return uint64(n), err
}

// Send a tx to the local server
func (mod *MonkRpcModule) rpcLocalTxCall(args monkrpc.NewTxArgs) (string, error) {
	return mod.rpcTxCall("Transact", args)
}

// Send a tx to the remote server
func (mod *MonkRpcModule) rpcRemoteTxCall(args monkrpc.PushTxArgs) (string, error) {
	return mod.rpcTxCall("PushTx", args)
}

func (mod *MonkRpcModule) rpcTxCall(method string, args interface{}) (string, error) {
	res := new(string)
	err := mod.client.Call("TheloniousApi."+method, args, res)
	if err != nil {
		return "", err
	}
	return *res, nil
}

func NewDatabase(dbName string) monkutil.Database {
	db, err := monkdb.NewLDBDatabase(dbName)
	if err != nil {
		exit(err)
	}
	return db
}

func NewKeyManager(KeyStore string, Datadir string, db monkutil.Database) *monkcrypto.KeyManager {
	var keyManager *monkcrypto.KeyManager
	switch {
	case KeyStore == "db":
		keyManager = monkcrypto.NewDBKeyManager(db)
	case KeyStore == "file":
		keyManager = monkcrypto.NewFileKeyManager(Datadir)
	default:
		exit(fmt.Errorf("unknown keystore type: %s", KeyStore))
	}
	return keyManager
}

func exit(err error) {
	status := 0
	if err != nil {
		fmt.Println(err)
		status = 1
	}
	os.Exit(status)
}

func InitLogging(Datadir string, LogFile string, LogLevel int, DebugFile string) {
	var writer io.Writer
	if LogFile == "" {
		writer = os.Stdout
	} else {
		writer = openLogFile(Datadir, LogFile)
	}
	monklog.AddLogSystem(monklog.NewStdLogSystem(writer, log.LstdFlags, monklog.LogLevel(LogLevel)))
	if DebugFile != "" {
		writer = openLogFile(Datadir, DebugFile)
		monklog.AddLogSystem(monklog.NewStdLogSystem(writer, log.LstdFlags, monklog.DebugLevel))
	}
}

func AbsolutePath(Datadir string, filename string) string {
	if path.IsAbs(filename) {
		return filename
	}
	return path.Join(Datadir, filename)
}

func openLogFile(Datadir string, filename string) *os.File {
	path := AbsolutePath(Datadir, filename)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening log file '%s': %v", filename, err))
	}
	return file
}

// compile LLL file into evm bytecode
// returns hex
func CompileLLL(filename string, literal bool) string {
	code, err := monkutil.CompileLLL(filename, literal)
	if err != nil {
		fmt.Println("error compiling lll!", err)
		return ""
	}
	return monkutil.Bytes2Hex(code)
}
