package modules

type NoArgs struct {
}

type VString struct {
	SVal string
}

type VStringArr struct {
	Svals []string
}

type VBool struct {
	BVal bool
}

type VInteger struct {
	IVal int
}


type Account struct{
    Address string
    Balance string
    Nonce string
    Script string
    Storage *Storage

    IsScript bool
}

// Ordered map for storage in an account or generalized table
type Storage struct {
	// hex strings for eth, arrays of strings (cols) for sql dbs
	Storage map[string]string
	Order   []string
}

// Ordered map for all accounts
type State struct {
	State map[string]*Storage // map addrs to map of storage to value
	Order []string           // ordered addrs and ordered storage inside
}

type WorldState struct{
    Accounts map[string] *Account
    Order []string
}

type BlockMiniData struct {
	Number           string
	Hash             string
	Transactions     int
	PrevHash         string
	AccountsAffected []*AccountMini
}

type StateAtArgs struct {
	Address string
	Storage string
}

type TransactionArgs struct {
	BlockHash string
	TxHash    string
}

type Block struct {
	Number       string
	Time         int
	Nonce        string
	Hash         string
	PrevHash     string
	Difficulty   string
	Coinbase     string
	Transactions []*Transaction
	Uncles       []string
	GasLimit     string
	GasUsed      string
	MinGasPrice  string
	TxRoot        string
	UncleRoot     string
}

type Transaction struct {
	ContractCreation bool
	Nonce            string
	Hash             string
	Sender           string
	Recipient        string
	Value            string
	Gas              string
	GasCost          string
	BlockHash        string
	Error            string
}

type TxIndata struct {
	Recipient string
	Gas       string
	GasCost   string
	Value     string
	// endline is the separator for tx data. Each string is padded with 0's to become 32 bytes.
	Data      string
}

type TxReceipt struct {
	Success  bool   // If transaction hash was created basically.
	Compiled bool   // If a contract was created, and the txdata was successfully compiled.
	Address  string // If a contract was created.
	Hash     string // Transaction hash
	Error    string
}

type AccountMini struct {
	// Modified (0), Added (1), Deleted(2)
	Flag     int
	Contract bool
	Address  string
	Nonce    string
	Balance  string
}