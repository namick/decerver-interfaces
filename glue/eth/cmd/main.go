package main

import (
    "time"
    "github.com/eris-ltd/decerver-interfaces/glue/eth"
)

func main(){

    e := eth.NewEth(nil)
    e.Init()
    e.Start()
    time.Sleep(10*time.Second)

}
