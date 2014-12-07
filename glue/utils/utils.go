package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
    "bytes"
    "io/ioutil"
    "encoding/json"

	"github.com/eris-ltd/thelonious/monkcrypto"
	"github.com/eris-ltd/thelonious/monkdb"
	"github.com/eris-ltd/thelonious/monklog"
	"github.com/eris-ltd/thelonious/monkutil"
	"github.com/eris-ltd/thelonious/monkwire"
)

var (
	GoPath = os.Getenv("GOPATH")
	usr, _ = user.Current() // error?!
    ErisLtd = path.Join(GoPath, "src", "github.com", "eris-ltd")

    Decerver = path.Join(usr.HomeDir, ".decerver")
    Apps = path.Join(Decerver, "apps")    
    Blockchains = path.Join(Decerver, "blockchains")    
    Filesystems = path.Join(Decerver, "filesystems")    
    Logs = path.Join(Decerver, "logs")
    Modules = path.Join(Decerver, "modules")
    Scratch = path.Join(Decerver, "scratch")
)

func NewDatabase(dbName string) monkutil.Database {
	db, err := monkdb.NewLDBDatabase(dbName)
	if err != nil {
		exit(err)
	}
	return db
}

func NewClientIdentity(clientIdentifier, version, customIdentifier string) *monkwire.SimpleClientIdentity {
	return monkwire.NewSimpleClientIdentity(clientIdentifier, version, customIdentifier)
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
	if !monklog.IsNil() {
		return
	}
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

// common golang, really?
func Copy(src, dst string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	return nil
}

func InitDataDir(Datadir string) error {
	_, err := os.Stat(Datadir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Creating directory: '%s'\n", Datadir)
			err := os.MkdirAll(Datadir, 0777)
            if err != nil{
                return err
            }
		}
	}
    return nil
}

func InitDecerverDir() error{
    dirs := []string{Decerver, Apps, Blockchains, Filesystems, Logs, Modules, Scratch}
    for _, d := range dirs{
        err := InitDataDir(d)
        if err != nil{
            return err
        }
    }
    err := InitDataDir(path.Join(Blockchains, "refs"))
    if err != nil{
        return err
    }
    _, err = os.Create(path.Join(Blockchains, "HEAD"))
    return err
}

func WriteJson(config interface{}, config_file string) error{
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
    if err != nil{
        return err
    }
	err = ioutil.WriteFile(config_file, out.Bytes(), 0600)
    return err
}

func ChainIdFromName(name string) string{
    b, err := ioutil.ReadFile(path.Join(Blockchains, "refs", name))
    if err != nil{
        return ""
    }
    return string(b)
}

func NewChainRef(name, chainId string) error{
    p := path.Join(Blockchains, "refs", name)
    _, err := os.Stat(p)
    if err == nil{
        return fmt.Errorf("Chain named %s already exists", name)
    }
    return ioutil.WriteFile(p, []byte(chainId), 0644)
}


