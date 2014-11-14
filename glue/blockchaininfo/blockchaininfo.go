package blockchaininfo

import (
    "net/http"
    "path"
    "io/ioutil"
    "encoding/json"
    "strconv"
    "log"
    "fmt"

	"github.com/eris-ltd/decerver-interfaces/api"
	"github.com/eris-ltd/decerver-interfaces/core"
	"github.com/eris-ltd/decerver-interfaces/events"
    "github.com/eris-ltd/decerver-interfaces/modules"

    "github.com/qedus/blockchain"
)

type BlkChainInfo struct {

    BciApi *blockchain.BlockChain
    Addresses *modules.Addresses
    config string
	chans map[string]chan events.Event

}

func NewBlkChainInfo() *BlkChainInfo{
    return &BlkChainInfo{}
}

/*

    module functions to satisfy interface. see:
        * https://github.com/eris-ltd/decerver-interfaces/blob/master/modules/modules.go

*/
func (b *BlkChainInfo) Register(fileIO core.FileIO, registry api.ApiRegistry, runtime core.Runtime, eReg events.EventRegistry) error {
    b.config = path.Join(fileIO.Modules(), "blockchain", "config")
    return nil
}

func (b *BlkChainInfo) Init() error {
    bc := blockchain.New(http.DefaultClient)
    b.BciApi = bc
    b.Addresses = &modules.Addresses{}

    cfg, err := ioutil.ReadFile(b.config)
    if err != nil {
        return err
    }

    bciCfg := make(map[string]string)
    err = json.Unmarshal(cfg, bciCfg)
    if err != nil {
        return err
    }

    b.BciApi.GUID           = bciCfg["guid"]
    b.BciApi.Password       = bciCfg["password"]
    b.BciApi.SecondPassword = bciCfg["second_password"]
    b.BciApi.APICode        = bciCfg["api_code"]

    var a1 *blockchain.AddressList
    if b.BciApi.GUID != "" {
        a1 = &blockchain.AddressList{}
        if err := bc.Request(a1); err != nil {
            return err
        }
    }
    bciAccountListToDecerverAccountList(a1, b.Addresses)

    b.chans = make(map[string]chan events.Event)
	return nil
}

// Start is not supported
// todo - start the polling and default subscribers
func (b *BlkChainInfo) Start() error {
    return nil
}

// Shutdown is not supported
// todo - unsubscribe and stop the long polling
func (b *BlkChainInfo) Shutdown() error {
    return nil
}

func (b *BlkChainInfo) Name() string {
	return "blockchaininfo"
}

/*

    blockchain functions to satisfy interface. see:
        * https://github.com/eris-ltd/decerver-interfaces/blob/master/modules/blockchain.go
        * https://github.com/eris-ltd/decerver-interfaces/blob/master/modules/modules.go

*/
// WorldState is not supported
func (b *BlkChainInfo) WorldState() *modules.WorldState {
    return &modules.WorldState{}
}

// State is not supported
func (b *BlkChainInfo) State() *modules.State {
    return &modules.State{}
}

// Storage is not supported
func (b *BlkChainInfo) Storage(target string) *modules.Storage {
    return &modules.Storage{}
}

// Account ...
func (b *BlkChainInfo) Account(target string) *modules.Account {
    a1 := &blockchain.Address{Address: target}
    if err := b.BciApi.Request(a1); err != nil {
        log.Print(err)
    }
    a2 := &modules.Account{}
    bciAccountToDecerverAccount(a1, a2)
    return a2
}

// StorageAt is not supported by this module
func (b *BlkChainInfo) StorageAt(target, storage string) string {
    return ""
}

// BlockCount ...
func (b *BlkChainInfo) BlockCount() int {
    block := &blockchain.LatestBlock{}
    if err := b.BciApi.Request(block); err != nil {
        log.Print(err)
    }
    return int(block.Height)
}

// LatestBlock ...
func (b *BlkChainInfo) LatestBlock() string {
    block := &blockchain.LatestBlock{}
    if err := b.BciApi.Request(block); err != nil {
        log.Print(err)
    }
    return block.Hash
}

// Block ...
func (b *BlkChainInfo) Block(hash string) *modules.Block {
    b1 := &blockchain.Block{Hash: hash}
    if err := b.BciApi.Request(b1); err != nil {
        log.Print(err)
    }
    b2 := &modules.Block{}
    bciBlocksToDecerverBlocks(b1, b2)
    return b2
}

// IsScript will always return false as no target address on the BTC chain will be a script address
func (b *BlkChainInfo) IsScript(target string) bool {
    return false
}

// Tx ...
func (b *BlkChainInfo) Tx(addr, amt string) (string, error) {
    amtt, err := strconv.Atoi(amt)
    if err != nil {
        return "", err
    }
    sp := &blockchain.SendPayment{
        Amount:    int64(amtt),
        ToAddress: addr
    }
    if err = b.BciApi.Request(sp); err != nil {
        return "", err
    } else {
        return sp.TransactionHash, nil
    }
    return "", nil
}

// Msg not supported by this module which is an API Wrapper around Blockchain.info
func (b *BlkChainInfo) Msg(addr string, data []string) (string, error) {
    return "", nil
}

// Script not supported by this module which is an API Wrapper around Blockchain.info
func (b *BlkChainInfo) Script(file, lang string) (string, error) {
    return "", nil
}

// todo
func (b *BlkChainInfo) Subscribe(name, event, target string) chan events.Event {
    ch := make(chan events.Event)
    return ch
}

// todo
func (b *BlkChainInfo) UnSubscribe(name string) {

}

// Commit not supported by this module which is an API Wrapper around Blockchain.info
func (b *BlkChainInfo) Commit() {}

// AutoCommit not supported by this module which is an API Wrapper around Blockchain.info
func (b *BlkChainInfo) AutoCommit(toggle bool) {}

// IsAutocommit not supported by this module which is an API Wrapper around Blockchain.info
func (b *BlkChainInfo) IsAutocommit() bool {
    return false
}

/*

    keymanager functions to satisfy interface. see:
        * https://github.com/eris-ltd/decerver-interfaces/blob/master/modules/blockchain.go
        * https://github.com/eris-ltd/decerver-interfaces/blob/master/modules/modules.go

*/
func (b *BlkChainInfo) ActiveAddress() string {
    return b.Addresses.ActiveAddress
}

func (b *BlkChainInfo) Address(n int) (string, error) {
    if b.Addresses.AddressList[n] != "" {
        return b.Addresses.AddressList[n], nil
    } else {
        return "", fmt.Errorf("Address does not exist at that index.")
    }
    return "", nil
}

func (b *BlkChainInfo) SetAddress(addr string) error {
    for _, add := range b.Addresses.AddressList {
        if addr == add {
            b.Addresses.ActiveAddress = addr
            return nil
        }
    }
    return fmt.Errorf("Requested address does not exist in Address List.")
}

func (b *BlkChainInfo) SetAddressN(n int) error {
    if (n >= len(b.Addresses.AddressList)) {
        return fmt.Errorf("Address does not exist at that index.")
    }
    b.Addresses.ActiveAddress = b.Addresses.AddressList[n]
    return nil
}

func (b *BlkChainInfo) NewAddress(set bool) string {
    na := &blockchain.NewAddress{Label: "via-decerver"}
    if err := b.BciApi.Request(na); err != nil {
        log.Print(err)
    }
    if set {
        b.SetAddress(na.Address)
    }
    return na.Address
}

func (b *BlkChainInfo) AddressCount() int {
    return len(b.Addresses.AddressList)
}

/*

    helper functions

*/
func bciAccountListToDecerverAccountList(a1 *blockchain.AddressList, a2 *modules.Addresses) {
    for _, add := range a1.Addresses {
        a2.AddressList = append(a2.AddressList, add.Address)
    }
}

func bciBlocksToDecerverBlocks(b1 *blockchain.Block, b2 *modules.Block) {
    b2.Number   = strconv.Itoa(int(b1.Height))
    b2.Time     = int(b1.Time)
    b2.Hash     = b1.Hash
    b2.PrevHash = b1.PreviousBlock
    b2.Nonce    = strconv.Itoa(int(b1.Nonce))
    b2.TxRoot   = b1.MerkelRoot

    for i := range b1.Transactions {
        b2.Transactions = append(b2.Transactions, &modules.Transaction{})
        bciTxToDecerverTx(&b1.Transactions[i], b2.Transactions[i])
    }
}

func bciTxToDecerverTx(t1 *blockchain.Transaction, t2 *modules.Transaction) {
    t2.Hash = t1.Hash
    for i := range t1.Inputs {
        t2.Inputs = append(t2.Inputs, &modules.Input{})
        bciInputsToDecerverInputs(&t1.Inputs[i], t2.Inputs[i])
    }

    for i := range t1.Outputs {
        t2.Outputs = append(t2.Outputs, &modules.Output{})
        bciOutputsToDecerverOutputs(&t1.Outputs[i], t2.Outputs[i])
    }
}

func bciInputsToDecerverInputs(i1 *blockchain.Input, i2 *modules.Input) {
    i2.PrevOut.Address = i1.PrevOut.Address
    i2.PrevOut.Number  = i1.PrevOut.Number
    i2.PrevOut.Type    = i1.PrevOut.Type
    i2.PrevOut.Value   = i1.PrevOut.Value
}

func bciOutputsToDecerverOutputs(o1 *blockchain.Output, o2 *modules.Output) {
    o2.Address = o1.Address
    o2.Number  = o1.Number
    o2.Type    = o1.Type
    o2.Value   = o1.Value
}

func bciAccountToDecerverAccount(a1 *blockchain.Address, a2 *modules.Account) {
    a2.Address  = a1.Address
    a2.Balance  = strconv.Itoa(int(a1.FinalBalance))
    a2.Nonce    = strconv.Itoa(int(a1.TransactionCount))
    a2.IsScript = false
}
