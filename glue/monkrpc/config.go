package monkrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	mutils "github.com/eris-ltd/decerver-interfaces/glue/monkutils"
	"github.com/eris-ltd/decerver-interfaces/glue/utils"
	"github.com/eris-ltd/thelonious/monkutil"
	"io/ioutil"
	"os"
	"path"
	"reflect"
)

var ErisLtd = utils.ErisLtd

type RpcConfig struct {
	// Networking
	RpcHost string `json:"rpc_host"`
	RpcPort int    `json:"rpc_port"`

	// If true, key management is handled
	// by the server (presumably on a local machine)
	// else, txs are signed by a key and rlp serialized
	Local bool `json:"local"`

	// Only relevant if Local is false
	KeySession string `json:"key_session"`
	KeyStore   string `json:"key_store"`
	KeyCursor  int    `json:"key_cursor"`
	KeyFile    string `json:"key_file"`

	// Paths
	RootDir      string `json:"root_dir"`
	DbName       string `json:"db_name"`
	LLLPath      string `json:"lll_path"`
	ContractPath string `json:"contract_path"`

	// Logs
	LogFile   string `json:"log_file"`
	DebugFile string `json:"debug_file"`
	LogLevel  int    `json:"log_level"`
}

// set default config object
var DefaultConfig = &RpcConfig{
	// Network
	RpcHost: "",
	RpcPort: 30304,

	Local: true,

	// Local Node
	KeySession: "generous",
	KeyStore:   "file",
	KeyCursor:  0,
	KeyFile:    path.Join(ErisLtd, "thelonious", "monk", "keys.txt"),

	// Paths
	RootDir:      path.Join(usr.HomeDir, ".monkchain2"),
	DbName:       "database",
	LLLPath:      "NETCALL", //path.Join(homeDir(), "cpp-ethereum/build/lllc/lllc"),
	ContractPath: path.Join(ErisLtd, "eris-std-lib"),

	// Log
	LogFile:   "",
	DebugFile: "",
	LogLevel:  5,
}

// Marshal the current configuration to file in pretty json.
func (mod *MonkRpcModule) WriteConfig(config_file string) {
	b, err := json.Marshal(mod.Config)
	if err != nil {
		fmt.Println("error marshalling config:", err)
		return
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")
	ioutil.WriteFile(config_file, out.Bytes(), 0600)
}

// Unmarshal the configuration file into module's config struct.
func (mod *MonkRpcModule) ReadConfig(config_file string) {
	b, err := ioutil.ReadFile(config_file)
	if err != nil {
		fmt.Println("could not read config", err)
		fmt.Println("resorting to defaults")
		mod.WriteConfig(config_file)
		return
	}
	var config RpcConfig
	err = json.Unmarshal(b, &config)
	if err != nil {
		fmt.Println("error unmarshalling config from file:", err)
		fmt.Println("resorting to defaults")
		return
	}
	*(mod.Config) = config
}

// Set a field in the config struct.
func (mod *MonkRpcModule) SetConfig(field string, value interface{}) error {
	cv := reflect.ValueOf(mod.Config).Elem()
	f := cv.FieldByName(field)
	kind := f.Kind()

	k := reflect.ValueOf(value).Kind()
	if kind != k {
		return fmt.Errorf("Invalid kind. Expected %s, received %s", kind, k)
	}

	if kind == reflect.String {
		f.SetString(value.(string))
	} else if kind == reflect.Int {
		f.SetInt(int64(value.(int)))
	} else if kind == reflect.Bool {
		f.SetBool(value.(bool))
	}
	return nil
}

// Set the config object directly
func (mod *MonkRpcModule) SetConfigObj(config interface{}) error {
	if c, ok := config.(*RpcConfig); ok {
		mod.Config = c
	} else {
		return fmt.Errorf("Invalid config object")
	}
	return nil
}

// Set package global variables (LLLPath, monkutil.Config, logging).
// Create the root data dir if it doesn't exist, and copy keys if they are available
func (mod *MonkRpcModule) rConfig() {
	cfg := mod.Config
	// set lll path
	if cfg.LLLPath != "" {
		monkutil.PathToLLL = cfg.LLLPath
	}

	// check on data dir
	// create keys
	_, err := os.Stat(cfg.RootDir)
	if err != nil {
		os.Mkdir(cfg.RootDir, 0777)
		_, err := os.Stat(path.Join(cfg.RootDir, cfg.KeySession) + ".prv")
		if err != nil {
			utils.Copy(cfg.KeyFile, path.Join(cfg.RootDir, cfg.KeySession)+".prv")
		}
	}
	// a global monkutil.Config object is used for shared global access to the db.
	// this also uses rakyl/globalconf, but we mostly ignore all that
	if monkutil.Config == nil {
		monkutil.Config = &monkutil.ConfigManager{ExecPath: cfg.RootDir, Debug: true, Paranoia: true}
	}

	if monkutil.Config.Db == nil {
		monkutil.Config.Db = mutils.NewDatabase(mod.Config.DbName)
	}

	// TODO: enhance this with more pkg level control
	utils.InitLogging(cfg.RootDir, cfg.LogFile, cfg.LogLevel, cfg.DebugFile)
}
