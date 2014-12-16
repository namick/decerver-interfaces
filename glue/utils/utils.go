package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/eris-ltd/decerver-interfaces/dapps"
	"github.com/eris-ltd/thelonious/monklog"
)

var (
	GoPath  = os.Getenv("GOPATH")
	ErisLtd = path.Join(GoPath, "src", "github.com", "eris-ltd")

	usr, _      = user.Current() // error?!
	Decerver    = path.Join(usr.HomeDir, ".decerver")
	Apps        = path.Join(Decerver, "dapps")
	Blockchains = path.Join(Decerver, "blockchains")
	Filesystems = path.Join(Decerver, "filesystems")
	Logs        = path.Join(Decerver, "logs")
	Modules     = path.Join(Decerver, "modules")
	Scratch     = path.Join(Decerver, "scratch")
	HEAD        = path.Join(Blockchains, "HEAD")
	Refs        = path.Join(Blockchains, "refs")
	Epm         = path.Join(Scratch, "epm")
	Lllc        = path.Join(Scratch, "lllc")
)

var MajorDirs = []string{
	Decerver, Apps, Blockchains, Filesystems, Logs, Modules, Scratch, Refs, Epm, Lllc,
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
			err := os.MkdirAll(Datadir, 0777)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func InitDecerverDir() error {
	for _, d := range MajorDirs {
		err := InitDataDir(d)
		if err != nil {
			return err
		}
	}
	err := InitDataDir(Refs)
	if err != nil {
		return err
	}
	if _, err = os.Stat(HEAD); err != nil {
		_, err = os.Create(HEAD)
	}
	return err
}

func WriteJson(config interface{}, config_file string) error {
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(config_file, out.Bytes(), 0600)
	return err
}

func ChainIdFromName(name string) string {
	refsPath := path.Join(Blockchains, "refs", name)
	b, err := ioutil.ReadFile(refsPath)
	if err != nil {
		return ""
	}
	return string(b)
}

func CheckGetPackageFile(dappDir string) (*dapps.PackageFile, error) {
	if _, err := os.Stat(dappDir); err != nil {
		return nil, fmt.Errorf("Dapp %s not found", dappDir)
	}

	b, err := ioutil.ReadFile(path.Join(dappDir, "package.json"))
	if err != nil {
		return nil, err
	}

	p, err := dapps.NewPackageFileFromJson(b)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ChainIdFromDapp(dapp string) (string, error) {
	p, err := CheckGetPackageFile(path.Join(Apps, dapp))
	if err != nil {
		return "", err
	}

	var chainId string
	for _, dep := range p.ModuleDependencies {
		if dep.Name == "monk" {
			d := &dapps.MonkData{}
			if err := json.Unmarshal(dep.Data, d); err != nil {
				return "", err
			}
			chainId = d.ChainId
		}
	}
	if chainId == "" {
		return "", fmt.Errorf("Dapp is missing monk dependency or chainId!")
	}

	return chainId, nil
}

func ResolveChainType(chainType string) string {
	switch chainType {
	case "thel", "thelonious", "monk":
		return "thelonious"
	case "btc", "bitcoin":
		return "bitcoin"
	case "eth", "ethereum":
		return "ethereum"
	case "gen", "genesis":
		return "thelonious"
	}
	return ""
}

// Determines the chainId from a chainId prefix or from a ref, but not from a dapp
func ResolveChainId(chainType, name, chainId string) (string, error) {
	chainType = ResolveChainType(chainType)
	if chainType == "" {
		return "", fmt.Errorf("Unknown chain type: ", chainType)
	}

	var p string
	idFromName := ChainIdFromName(name)
	if idFromName != "" {
		chainId = idFromName
	}

	if chainId != "" {
		p = path.Join(Blockchains, chainType, chainId)
		if _, err := os.Stat(p); err != nil {
			// see if its a prefix of a chainId
			id, err := findPrefixMatch(path.Join(Blockchains, chainType), chainId)
			if err != nil {
				return "", err
			}
			p = path.Join(Blockchains, chainType, id)
			chainId = id
		}
	}
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("Could not locate chain by name %s or by id %s", name, chainId)
	}

	return chainId, nil

}

func ResolveChain(chainType, name, chainId string) (string, error) {
	id, err := ResolveChainId(chainType, name, chainId)
	if err != nil {
		return "", err
	}
	return path.Join(Blockchains, chainType, id), nil
}

func findPrefixMatch(dirPath, prefix string) (string, error) {
	fs, _ := ioutil.ReadDir(dirPath)
	found := false
	var p string
	for _, f := range fs {
		if strings.HasPrefix(f.Name(), prefix) {
			if found {
				return "", fmt.Errorf("ChainId collision! Multiple chains begin with %s. Please be more specific", prefix)
			}
			p = f.Name() //path.Join(Blockchains, chainType, f.Name())
			found = true
		}
	}
	if !found {
		return "", fmt.Errorf("ChainId %s did not match any known chains", prefix)
	}
	return p, nil
}

// Maximum entries in the HEAD file
var MaxHead = 100

// The HEAD file is a running list of the latest head
// so we can go back if we mess up or forget
func ChangeHead(head string) error {
	head, err := ResolveChainId("thelonious", head, head)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(HEAD)
	if err != nil {
		return err
	}
	bspl := strings.Split(string(b), "\n")
	var bsp string
	if len(bspl) >= MaxHead {
		bsp = strings.Join(bspl[:MaxHead-1], "\n")
	} else {
		bsp = string(b)
	}
	bsp = head + "\n" + bsp
	err = ioutil.WriteFile(HEAD, []byte(bsp), 0666)
	if err != nil {
		return err
	}
	return nil
}

// Add a reference name to a chainId
func AddRef(id, ref string) error {
	_, err := os.Stat(path.Join(Refs, ref))
	if err == nil {
		return fmt.Errorf("Ref %s already exists", ref)
	}

	dataDir := path.Join(Blockchains, "thelonious")
	_, err = os.Stat(path.Join(dataDir, id))
	if err != nil {
		id, err = findPrefixMatch(dataDir, id)
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(path.Join(Refs, ref), []byte(id), 0644)
}

// Return a list of chain references
func GetRefs() (map[string]string, error) {
	fs, err := ioutil.ReadDir(Refs)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, f := range fs {
		name := f.Name()
		b, err := ioutil.ReadFile(path.Join(Refs, name))
		if err != nil {
			return nil, err
		}
		m[name] = string(b)
	}
	return m, nil
}

// Get the current active chain
func GetHead() (string, error) {
	// TODO: only read the one line!
	f, err := ioutil.ReadFile(HEAD)
	if err != nil {
		return "", err
	}
	fspl := strings.Split(string(f), "\n")
	return fspl[0], nil
}
