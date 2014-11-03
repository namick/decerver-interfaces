package core

import (
	"os"
)

type FileIO interface {
	Root() string
	Databases() string
	FileSystems() string
	Log() string
	Apps() string
	OpenFile(path string, name string) (*os.File, error)
	SaveFile(path string, name string) error
}

type Runtime interface {
	BindScriptObject(name string, val interface{}) error
	LoadScriptFile(fileName string) error
	LoadScriptFiles(fileName ...string) error
	RunAction(path []string, actionName string, params interface{}) ([]string, error)
	RunMethod(nameSpace, funcName string, params interface{}) ([]string, error)
}
