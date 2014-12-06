package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	_ "strconv"

	"github.com/eris-ltd/thelonious/monkcrypto"
	"github.com/eris-ltd/thelonious/monkdb"
	"github.com/eris-ltd/thelonious/monklog"
	"github.com/eris-ltd/thelonious/monkutil"
	"github.com/eris-ltd/thelonious/monkwire"
)

var (
	GoPath = os.Getenv("GOPATH")
	usr, _ = user.Current() // error?!
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

func InitDataDir(Datadir string) {
	_, err := os.Stat(Datadir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Data directory '%s' doesn't exist, creating it\n", Datadir)
			os.Mkdir(Datadir, 0777)
		}
	}
}
