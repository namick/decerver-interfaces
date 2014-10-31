package core

import (
	"os"
	"github.com/robertkrimen/otto"
)

type FileSys interface {
	Open(path string, name string) (*os.File, error)
	Save(path string, name string) error
}

// A function that can be added to the otto vm.
type AteFunc func(otto.FunctionCall) otto.Value

type ScriptEngine interface {
	InjectFunction(funcName string, fun AteFunc)
	LoadScript(fileName string)
	RunAction(path []string, actionName string, params interface{}) []string
	RunMethod(nameSpace, funcName string, params interface{}) []string
}