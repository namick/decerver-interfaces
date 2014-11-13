package main

import (
    "fmt"
    "log"
    "time"
    "encoding/hex"
    "github.com/eris-ltd/glululemon/ipfs"
    "github.com/eris-ltd/decerver-interfaces/modules"
)


func main(){
    i := ipfs.NewIpfs()
    err := i.Init()
    if err != nil{
        log.Fatal(err)
    }
    start := time.Now()
    err = i.Start()   
    if err != nil{
        log.Fatal(err)
    }
    fmt.Println("startup took:", time.Since(start))

    //c := "QmVHdqmE5x55kZaavWUmscLmieusDdZhQBP5mjZHwMB3U9"
    c := "Qmb8zwr341xu5uUWwxvVKbZs1ZbjJRJJ965tnV9HDeVUkH"
    h, _ := ipfs.B58ToHex(c)
    /*
    a, _ := i.Get("file", h)
    fmt.Println(string(a.([]byte)))

    g, _ := i.Get("tree", ipfs.B58ToHex("QmaKxiCScMY6BG1eq228F2fDJmjxZ53MJ8MtEyEJZr3v44"))
    t := g.(modules.FsNode)
    printTree(&t)
    */

    ch, _ := i.Get("stream", h)
    for r := range ch.(chan []byte){
        fmt.Println(string(r))
    }

    a := hex.EncodeToString([]byte("fuck you"))
    fmt.Println("#####") 
    k, _ := i.Push("block", a) 
    fmt.Println(k)
    aa, err := i.Get("block", k)
    if err != nil{
        fmt.Println(err)
    }
    fmt.Println(string(aa.([]byte)))

}

func printTree(t *modules.FsNode){
    fmt.Println(t.Name)
    for _, tt := range t.Nodes{
        printTree(tt)
    }
}


