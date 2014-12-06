package main

import (
	"github.com/eris-ltd/decerver-interfaces/glue/eth"
	"time"
)

func main() {

	e := eth.NewEth(nil)
	e.Init()
	e.Start()
	time.Sleep(10 * time.Second)

}
