package core

import (
)

type DCConfig struct {
	RootDir    string `json:"deCerverDirectory"`
	LogFile    string `json:"logFile"`
	MaxClients int    `json:"maxClients"`
	Port       int    `json:"portNumber"`
}

type DeCerver interface {
	ReadConfig(filename string)
	WriteConfig(cfg *DCConfig)
	GetConfig() *DCConfig
	GetPaths() FileIO
}

type FileIO interface {
	Root() string
	Log() string
	Apps() string
	Blockchains() string
	Filesystems() string
	Modules() string
	// Useful when you want to load a file inside of a directory gotten by the 
	// 'Paths' object. Reads and returns the bytes.
	ReadFile(directory, name string) ([]byte, error)
	// Useful when you want to save a file into a directory gotten by the 'Paths'
	// object.
	WriteFile(directory, name string, data []byte) error
}

type Runtime interface {
	BindScriptObject(name string, val interface{}) error
	LoadScriptFile(fileName string) error
	LoadScriptFiles(fileName ...string) error
	AddScript(script string) error
	RunAction(path []string, actionName string, params interface{}) ([]string, error)
	CallFuncOnObj(objName, funcName string, params ...interface{})
	RunFunction(funcName string, params interface{}) ([]string, error)
}
