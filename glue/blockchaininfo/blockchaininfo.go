package blockchaininfo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/eris-ltd/decerver-interfaces/core"
	"github.com/eris-ltd/decerver-interfaces/events"
	"github.com/eris-ltd/decerver-interfaces/modules"

	"github.com/qedus/blockchain"
)

// BlkChainInfo is the main struct for the blockchain.info API module.
type BlkChainInfo struct {
	BciApi    *blockchain.BlockChain
	Addresses *modules.Addresses

	pollBlocks      chan bool
	mostRecentBlock string
	pollAddresses   chan bool
	addressesPolled map[string]string
	config          string
	chans           map[string]chan events.Event
}

// NewBlkChainInfo simply returns a pointer to a blank struct
func NewBlkChainInfo() *BlkChainInfo {
	return &BlkChainInfo{}
}

/*

   module functions to satisfy interface. see:
       * https://github.com/eris-ltd/decerver-interfaces/blob/master/modules/modules.go

*/

// Register sets the module config settings and returns nile
func (b *BlkChainInfo) Register(fileIO core.FileIO, rm core.RuntimeManager, eReg events.EventRegistry) error {
	b.config = path.Join(fileIO.Modules(), "blockchain", "config")
	return nil
}

// Init (which is called after Register) is the main startup function for the blockchain.info API wrapper
// it first sets the default values the module requires, then reads the configuration file, and then
// uses those configuration settings from the file to establish the default values needed for non-query
// functions on the BlockChain.info API.
func (b *BlkChainInfo) Init() error {

	// set default values
	b.BciApi = blockchain.New(http.DefaultClient)
	b.Addresses = &modules.Addresses{}
	b.chans = make(map[string]chan events.Event)
	b.addressesPolled = make(map[string]string)

	// read the config file
	cfg, err := ioutil.ReadFile(b.config)
	if err != nil {
		return err
	}

	// use the config file to establish the right settings for the API wrapper
	bciCfg := make(map[string]string)
	err = json.Unmarshal(cfg, bciCfg)
	if err != nil {
		return err
	}
	b.BciApi.GUID = bciCfg["guid"]
	b.BciApi.Password = bciCfg["password"]
	b.BciApi.SecondPassword = bciCfg["second_password"]
	b.BciApi.APICode = bciCfg["api_code"]

	// sets the address list.
	var a1 *blockchain.AddressList
	if b.BciApi.GUID != "" {
		a1 = &blockchain.AddressList{}
		if err := b.BciApi.Request(a1); err != nil {
			return err
		}
	}
	bciAccountListToDecerverAccountList(a1, b.Addresses)

	// sets the channels map
	b.chans = make(map[string]chan events.Event)
	return nil
}

// Start does nothing. But it is needed to satisfy the blockchain module.
func (b *BlkChainInfo) Start() error {
	return nil
}

// Shutdown simply stops any pollers
func (b *BlkChainInfo) Shutdown() error {
	b.stopPollBlocks()

	for addr := range b.addressesPolled {
		b.stopPollAddresses(addr)
	}

	return nil
}

// Name returns the name of the module: "blockchaininfo"
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

// Account queries the address passed to it, and translates the received object from the API
// wrapper into an appropriate struct to be consumed by the deCerver.
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

// BlockCount returns the block Height which blockchain.info reports
func (b *BlkChainInfo) BlockCount() int {
	block := &blockchain.LatestBlock{}
	if err := b.BciApi.Request(block); err != nil {
		log.Print(err)
	}
	return int(block.Height)
}

// LatestBlock returns the hash of the most recent block
func (b *BlkChainInfo) LatestBlock() string {
	block := &blockchain.LatestBlock{}
	if err := b.BciApi.Request(block); err != nil {
		log.Print(err)
	}
	return block.Hash
}

// Block queries blockchain.info for a block by the blockhash and then translates the
// struct received from the API wrapper into a struct which can be consumed by the decerver as
// a normal blockchain module block struct.
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

// Tx sends a transfer. Note that if the user has two factor authentication on in their blockchain.info
// account, the blockchain.info API will not allow transactions.
func (b *BlkChainInfo) Tx(addr, amt string) (string, error) {
	amtt, err := strconv.Atoi(amt)
	if err != nil {
		return "", err
	}
	sp := &blockchain.SendPayment{
		Amount:    int64(amtt),
		ToAddress: addr,
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

// Subscribe establishes long polling functions for either "newBlock" or "addr", "tx"
// Either call will return a channel of events which the decerver can consume. The
// former Subscribe function will return a block object whenever a new block is found
// by long polling the API. The latter Subscribe function will return a transaction struct
// whenever the watched address sends or receives a transaction.
func (b *BlkChainInfo) Subscribe(name, event, target string) chan events.Event {
	ch := make(chan events.Event)
	switch name {
	case "newBlock":
		ch = b.startPollBlocks()
	case "addr":
		if event == "tx" {
			ch = b.startPollAddresses(target)
		}
	}
	return ch
}

// UnSubscribe either turns off the long polling for the new block or deletes the
// address passed to it from the address transaction poller.
func (b *BlkChainInfo) UnSubscribe(name string) {
	switch name {
	case "newBlock":
		b.stopPollBlocks()
	default:
		b.stopPollAddresses(name)
	}
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
	if n >= len(b.Addresses.AddressList) {
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
	b2.Number = strconv.Itoa(int(b1.Height))
	b2.Time = int(b1.Time)
	b2.Hash = b1.Hash
	b2.PrevHash = b1.PreviousBlock
	b2.Nonce = strconv.Itoa(int(b1.Nonce))
	b2.TxRoot = b1.MerkelRoot

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
	i2.PrevOut.Number = i1.PrevOut.Number
	i2.PrevOut.Type = i1.PrevOut.Type
	i2.PrevOut.Value = i1.PrevOut.Value
}

func bciOutputsToDecerverOutputs(o1 *blockchain.Output, o2 *modules.Output) {
	o2.Address = o1.Address
	o2.Number = o1.Number
	o2.Type = o1.Type
	o2.Value = o1.Value
}

func bciAccountToDecerverAccount(a1 *blockchain.Address, a2 *modules.Account) {
	a2.Address = a1.Address
	a2.Balance = strconv.Itoa(int(a1.FinalBalance))
	a2.Nonce = strconv.Itoa(int(a1.TransactionCount))
	a2.IsScript = false
}

func (b *BlkChainInfo) startPollBlocks() chan events.Event {
	interval, _ := time.ParseDuration("2m")
	ticker := time.NewTicker(interval)
	b.pollBlocks = make(chan bool)
	ch := make(chan events.Event)
	b.chans["newBlock"] = ch
	go b.pollBlock(ticker)
	return ch
}

func (b *BlkChainInfo) stopPollBlocks() {
	b.pollBlocks <- true
}

func (b *BlkChainInfo) pollBlock(ticker *time.Ticker) {
	fmt.Println("[blockchain.info mod] Starting New Block Poller.")
	b.mostRecentBlock = b.LatestBlock()
	var rec string
	for {
		select {
		case <-ticker.C:
			fmt.Println("[blockchain.info mod] Polling for new block.")
			rec = b.LatestBlock()
			if rec != b.mostRecentBlock {
				b.mostRecentBlock = rec
				b2 := b.Block(rec)
				eve := events.Event{
					Event:     "newBlock",
					Resource:  b2,
					Source:    b.Name(),
					TimeStamp: time.Now(),
				}
				b.chans["newBlock"] <- eve
				fmt.Printf("[blockchain.info mod] New Block: %s.\n", rec)
			} else {
				fmt.Println("[blockchain.info mod] No New Block.")
			}
		case <-b.pollBlocks:
			fmt.Println("[blockchain.info mod] Stopping New Block Poller.")
			ticker.Stop()
			break
		}
	}
}

func (b *BlkChainInfo) startPollAddresses(addr string) chan events.Event {
	interval, _ := time.ParseDuration("1m")
	ticker := time.NewTicker(interval)
	if b.pollAddresses == nil {
		b.pollAddresses = make(chan bool)
	}
	ch := make(chan events.Event)
	b.chans[addr] = ch
	b.addressesPolled[addr] = ""
	if len(b.addressesPolled) == 1 {
		go b.pollAddress(ticker)
	}
	return ch
}

func (b *BlkChainInfo) stopPollAddresses(addr string) {
	delete(b.addressesPolled, addr)
	if len(b.addressesPolled) == 0 {
		b.pollAddresses <- true
	}
}

func (b *BlkChainInfo) pollAddress(ticker *time.Ticker) {
	fmt.Println("[blockchain.info mod] Starting New Address Poller.")
	for addr := range b.addressesPolled {
		b.addressesPolled[addr] = b.Account(addr).Nonce
	}
	rec := make(map[string]string)
	for {
		select {
		case <-ticker.C:
			fmt.Println("[blockchain.info mod] Polling Address(es).")
			for addr := range b.addressesPolled {
				rec[addr] = b.Account(addr).Nonce
			}
			for addr := range b.addressesPolled {
				if rec[addr] != b.addressesPolled[addr] {
					b.addressesPolled[addr] = rec[addr]

					// get the tx object so we can send that over the Events
					t1 := &blockchain.Address{Address: addr}
					if err := b.BciApi.Request(t1); err != nil {
						fmt.Println(err)
					}
					t2 := &modules.Transaction{}
					bciTxToDecerverTx(&t1.Transactions[len(t1.Transactions)-1], t2)

					// set and send the event
					eve := events.Event{
						Event:     "addressChanged",
						Resource:  t2,
						Source:    b.Name(),
						TimeStamp: time.Now(),
					}
					b.chans[addr] <- eve
					fmt.Printf("[blockchain.info mod] New transaction found for address: %s (New Nonce: %s)\n", addr, b.addressesPolled[addr])
				} else {
					fmt.Println("[blockchain.info mod] No New transactions found for address: ", addr)
				}
			}
		case <-b.pollAddresses:
			fmt.Println("[blockchain.info mod] Stopping Address Poller.")
			ticker.Stop()
			break
		}
	}
}
