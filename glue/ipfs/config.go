package ipfs

import (
	"os/user"
    "path"
)

// get users home directory
func homeDir() string {
	usr, _ := user.Current()
	return usr.HomeDir
}

type FSConfig struct{
    RootDir string 
    LogLevel int
}

var DefaultConfig = &FSConfig{
    RootDir: path.Join(homeDir(), ".go-ipfs") ,
    LogLevel: 5,
}

var logLevels = map[int]string{
    0:"critical",
    1:"error",
    2:"warning",
    3:"notice",
    4:"info",
    5:"debug",
}

