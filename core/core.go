package core

import (
	"os"
    "log"
)

type FileSys interface {
	Open(path string, name string) (*os.File, error)
	Save(path string, name string) error
}

// Ordered map for storage in an account or generalized table
type Storage struct{
    // hex strings for eth, arrays of strings (cols) for sql dbs
    Storage map[string]interface{}
    Order []string
}

// Ordered map for all accounts
type State struct{
    State map[string]Storage// map addrs to map of storage to value
    Order []string // ordered addrs and ordered storage inside
}

type LogSystem log.Logger

