package blockchain

import (
	"fmt"
	"github.com/eris-ltd/deCerver-interfaces/modules"
)

const (
	ERR_NO_SUCH_BLOCK        = "NO SUCH BLOCK"
	ERR_NO_SUCH_TX           = "NO SUCH TX"
	ERR_NO_SUCH_ADDRESS      = "NO SUCH ADDRESS"
	ERR_STATE_NO_STORAGE     = "STATE NO STORAGE"
	ERR_MALFORMED_ADDRESS    = "MALFORMED ADDRESS"
	ERR_MALFORMED_BLOCK_HASH = "MALFORMED BLOCK HASH"
	ERR_MALFORMED_TX_HASH    = "MALFORMED TX HASH"
)

const (
	ACCOUNT_MODIFIED = iota
	ACCOUNT_CREATED
	ACCOUNT_DELETED

    ZeroHash160 = "00000000000000000000"
    ZerHash256 = "00000000000000000000000000000000"
)

func getBlockChain(chain modules.Blockchain) []*modules.BlockMiniData {
	lastNum := chain.BlockCount()
	ctr := int(lastNum)
	fmt.Printf("Last Block Number: %d\n", lastNum)
	blocks := make([]*modules.BlockMiniData, ctr+1)
	bHash := chain.LatestBlock()
	block := chain.Block(bHash)
	fmt.Printf("Current Block Number: %s\n", block.Number)
	bmd := &modules.BlockMiniData{}
	getBlockMiniWSFromBlock(bmd, block)
	blocks[ctr] = bmd
	fmt.Printf("Current Block Mini: %v\n", bmd)
	ctr--
	for ctr >= 0 {
		pHash := block.PrevHash
		block = chain.Block(pHash)
		fmt.Printf("Current Block Number: %s\n", block.Number)
		bmd := &modules.BlockMiniData{}
		getBlockMiniWSFromBlock(bmd, block)
		blocks[ctr] = bmd
		fmt.Printf("Current Block Mini: %v\n", bmd)
		ctr--
	}

	return blocks
}

// Used during world state generation, when we don't care about the transactions.
func getBlockMiniWSFromBlock(reply *modules.BlockMiniData, block *modules.Block){

	reply.Number = block.Number
	reply.Hash = block.Hash

	if block.Transactions != nil && len(block.Transactions) > 0 {
		size := len(block.Transactions)
		reply.Transactions = size
	} else {
		reply.Transactions = 0
	}
}

// Used in block updates from reactor, when we want account diffs along with the block data.
func getBlockMiniDataFromBlock(chain modules.Blockchain, reply *modules.BlockMiniData, block *modules.Block) {

	reply.Number = block.Number
	reply.Hash = block.Hash

	aa := make(map[string]int)
	size := len(block.Transactions)
	reply.Transactions = size

	// Just check who sender and receiver is. Receiver may be a contract
	// creation address or a transaction receiver; either way it's a valid
	// account.
	for _, tx := range block.Transactions {

		// Sender cannot be anything other then modified, which
		// does not change the flag. It can however be unset.
		if _, ok := aa[tx.Sender]; !ok {
			aa[tx.Sender] = ACCOUNT_MODIFIED
		}

		// This flag is used for the receiver (or creation address).
		rFlag := ACCOUNT_MODIFIED
		var receiver string
		if tx.ContractCreation {
			rFlag |= ACCOUNT_CREATED
			receiver = tx.Recipient
		} else {
			receiver = tx.Recipient
		}

		// Receiver
		if _, ok := aa[receiver]; !ok {
			aa[receiver] = rFlag
		} else {
			aa[receiver] |= rFlag
		}

	}

	// Coinbase
	cbAddr := block.Coinbase

	if _, ok := aa[cbAddr]; !ok {
		aa[cbAddr] = ACCOUNT_MODIFIED
	}

	reply.AccountsAffected = make([]*modules.AccountMini, len(aa))
	// For the final step, we check if all the affected contracts still exist. If any of
	// the contracts has been removed, we update the flag to DELETED.
	ctr := 0

	for addr, flag := range aa {
		// TODO really convert back and forth between bytes...
		//addrBytes, _ := hex.DecodeString(addr)
		//stObj := ethChain.Ethereum.BlockChain().CurrentBlock.State().GetStateObject(addrBytes)
        acc := chain.Account(addr)
		am := &modules.AccountMini{}
		if acc == nil {
			am.Address = addr
			am.Flag = ACCOUNT_DELETED
		} else {
			am.Address = addr
			am.Nonce = acc.Nonce
			am.Balance = acc.Balance
			am.Flag = flag
		}
		reply.AccountsAffected[ctr] = am
		ctr++
	}
	
	// Block PrevHash
	if block.PrevHash != "" && block.PrevHash != ZeroHash160 {
		reply.PrevHash = block.PrevHash
	}
}

/*
func createTx(chain Blockchain, recipient, valueStr, gasStr, gasPriceStr, scriptStr string, reply *TxReceipt) error {
	var contractCreation bool
	if len(recipient) == 0 {
		contractCreation = true
	}
	hash, _ := hex.DecodeString(recipient)
	fmt.Printf("Recipient: %x\n", hash)
	value := ethutil.Big(valueStr)
	gas := ethutil.Big(gasStr)
	gasPrice := ethutil.Big(gasPriceStr)
	var tx *ethchain.Transaction
	// Compile and assemble the given data
	if contractCreation {
		// TODO disabled for now. Mutan is going away. Only LLL and Solidity.
		var script []byte
		var err error
		if ethutil.IsHex(scriptStr) {
			script, err = hex.DecodeString(scriptStr)
			reply.Compiled = false
		} else {
			script, err = ethutil.Compile(scriptStr, false)
			reply.Compiled = true
		}
		if err != nil {

			return err
		}

		tx = ethchain.NewContractCreationTx(value, gas, gasPrice, script)
	} else {
		data := ethutil.StringToByteFunc(scriptStr, func(s string) (ret []byte) {
			slice := strings.Split(s, "\n")
			for _, dataItem := range slice {
				d := ethutil.FormatData(dataItem)
				ret = append(ret, d...)
			}
			return
		})

		tx = ethchain.NewTransactionMessage(hash, value, gas, gasPrice, data)
	}

	keyPair := ethChain.Ethereum.KeyManager().KeyPair()
	acc := ethChain.Ethereum.StateManager().TransState().GetOrNewStateObject(keyPair.Address())
	tx.Nonce = acc.Nonce
	acc.Nonce += 1
	ethChain.Ethereum.StateManager().TransState().UpdateStateObject(acc)

	tx.Sign(keyPair.PrivateKey)
	ethChain.Ethereum.TxPool().QueueTransaction(tx)

	// Now write
	if contractCreation {
		reply.Address = hex.EncodeToString(tx.CreationAddress())
		fmt.Printf("Contract addr %x", tx.CreationAddress())
	}

	reply.Hash = hex.EncodeToString(tx.Hash())
	reply.Success = true

	return nil
}
*/

func getAccountMiniFromAccount(am *modules.AccountMini, acc *modules.Account) {
	am.Address = acc.Address
	am.Contract = len(acc.Script) > 0
	am.Balance = acc.Balance
	am.Nonce = acc.Nonce
	return
}