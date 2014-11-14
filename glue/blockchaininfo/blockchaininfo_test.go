package blockchaininfo

import (
	"fmt"
	// "log"
	// "io/ioutil"
	// "os"
	// "strconv"
	"github.com/eris-ltd/decerver-interfaces/modules"
	"testing"
)

var (
	BlockChainInfo = start()
	blockHash      = "000000000000000016d65758ed8df787c3d490c569578d38d6db2ed4b56817f0"
	acct1          = "15v4EdEsnt367mgUdqSvbS7xExXTwKWoTo"
)

func start() *BlkChainInfo {
	b := NewBlkChainInfo()
	_ = b.Init()
	// normally this would not _ the err, but for testing purposes there is no config file.
	return b
}

func testBlockEquality(block *modules.Block) error {
	if block.Number != "329896" {
		return fmt.Errorf("Block number is not right. Expected: 329896, Got: %s", block.Number)
	}

	if block.Time != 1415922366 {
		return fmt.Errorf("Block time is not right. Expected: %v, Got: %v", 1322131230, block.Time)
	}

	if block.Nonce != "2245627664" {
		return fmt.Errorf("Block nonce is not right. Expected: %s, Got: %s", "2964215930", block.Nonce)
	}

	if block.Hash != "000000000000000016d65758ed8df787c3d490c569578d38d6db2ed4b56817f0" {
		return fmt.Errorf("Go Kill Yourself. The blockhash searched on does not equal the blockhash returned.")
	}

	if block.PrevHash != "0000000000000000168017e70167b30132ee606e99fbbfc6bf7d0dcb0388286c" {
		return fmt.Errorf("Block previous hash is not right. Expected: %s, Got: %s", "0000000000000000168017e70167b30132ee606e99fbbfc6bf7d0dcb0388286c", block.PrevHash)
	}

	if block.TxRoot != "d3cee9d795cdee08ea36aeee2c2a481b2beb092d12d345c01134ab21d48d910f" {
		return fmt.Errorf("Block previous hash is not right. Expected: %s, Got: %s", "d3cee9d795cdee08ea36aeee2c2a481b2beb092d12d345c01134ab21d48d910f", block.TxRoot)
	}
	return nil
}

func testTxEquality(tx *modules.Transaction) error {
	return nil
}

func TestBlock(t *testing.T) {
	block := BlockChainInfo.Block(blockHash)
	err := testBlockEquality(block)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLatestBlock(t *testing.T) {
	latestBlock := BlockChainInfo.LatestBlock()
	if len(latestBlock) != 64 {
		t.Fatal("Latest block hash is incorrect.")
	}
}

func TestBlockHeight(t *testing.T) {
	blockHeight := BlockChainInfo.BlockCount()
	if blockHeight <= 329917 {
		t.Fatal("Block height is incorrect.")
	}
}

func TestAccount(t *testing.T) {
	acct1Res := BlockChainInfo.Account(acct1)
	if acct1Res.Balance != "0" {
		t.Fatalf("Incorrect balance. Expected: %s, Got: %s. Check https://blockchain.info/address/15v4EdEsnt367mgUdqSvbS7xExXTwKWoTo first.", 0, acct1Res.Balance)
	}
	if acct1Res.Nonce != "2" {
		t.Fatalf("Incorrect nonce. Expected: %s, Got: %s. Check https://blockchain.info/address/15v4EdEsnt367mgUdqSvbS7xExXTwKWoTo first.", 2, acct1Res.Nonce)
	}
}
