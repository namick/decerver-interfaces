package ipfs

import (
	//    "bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	decore "github.com/eris-ltd/decerver-interfaces/core"
	events "github.com/eris-ltd/decerver-interfaces/events"
	"github.com/eris-ltd/decerver-interfaces/modules"
	blocks "github.com/jbenet/go-ipfs/blocks"
	commands "github.com/jbenet/go-ipfs/commands"
	config "github.com/jbenet/go-ipfs/config"
	core "github.com/jbenet/go-ipfs/core"
	cmds "github.com/jbenet/go-ipfs/core/commands"
	mdag "github.com/jbenet/go-ipfs/merkledag"
	uio "github.com/jbenet/go-ipfs/unixfs/io"
	ftpb "github.com/jbenet/go-ipfs/unixfs/pb"
	u "github.com/jbenet/go-ipfs/util"
	util "github.com/jbenet/go-ipfs/util"
	//b58 "github.com/jbenet/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-base58"
	context "github.com/jbenet/go-ipfs/Godeps/_workspace/src/code.google.com/p/go.net/context"
	proto "github.com/jbenet/go-ipfs/Godeps/_workspace/src/code.google.com/p/goprotobuf/proto"
	mh "github.com/jbenet/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-multihash"
)

var (
	StreamSize = 1024
)

// implements decerver-interface module
type IpfsModule struct {
	ipfs   *Ipfs
	Config *FSConfig
}

// implements file system
type Ipfs struct {
	node *core.IpfsNode
	cfg  *config.Config
}

func (mod *IpfsModule) Register(fileIO decore.FileIO, rm decore.RuntimeManager, eReg events.EventRegistry) error {
	rm.RegisterApi("ipfs", mod)
	return nil
}

func NewIpfs() *IpfsModule {
	ii := new(IpfsModule)
	i := new(Ipfs)
	ii.Config = DefaultConfig
	ii.ipfs = i
	return ii
}

func (mod *IpfsModule) Init() error {
	// config is RootDir/config

	filename, err := config.Filename(mod.Config.RootDir)
	if err != nil {
		return err
	}

	// load the config file
	// if non-existant, initialize ipfs
	// on the machine
	mod.ipfs.cfg, err = config.Load(filename)

	if err != nil {
		if strings.Contains(err.Error(), "init") {
			c := exec.Command("ipfs", "init", "-d="+mod.Config.RootDir)
			c.Stdout = os.Stdout
			err := c.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
		} else {
			return err
		}
		return mod.Init()
	}

	u.SetLogLevel("*", "critical") //logLevels[mod.Config.LogLevel])

	/*if err := updates.CliCheckForUpdates(cfg, filename); err != nil {
		return nil, err
	}*/
	return nil
}

func (mod *IpfsModule) Start() error {
	n, err := core.NewIpfsNode(mod.ipfs.cfg, mod.Config.Online) //config, online
	if err != nil {
		return err
	}

	mod.ipfs.node = n
	return nil
}

// TODO: UDP socket won't close
// https://github.com/jbenet/go-ipfs/issues/389
func (mod *IpfsModule) Shutdown() error {
	if n := mod.ipfs.node.Network; n != nil {
		n.Close()
	}
	mod.ipfs.node.Close()
	return nil
}

func (mod *IpfsModule) ReadConfig(config_file string) {
}

func (mod *IpfsModule) WriteConfig(config_file string) {
}

func (mod *IpfsModule) Name() string {
	return "ipfs"
}

func (mod *IpfsModule) Subscribe(name string, event string, target string) chan events.Event {
	return mod.ipfs.Subscribe(name, event, target)
}

func (mod *IpfsModule) UnSubscribe(name string) {
	mod.ipfs.UnSubscribe(name)
}

func (mod *IpfsModule) Get(cmd string, params ...string) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.Get(cmd, params...))
}

func (mod *IpfsModule) Push(cmd string, params ...string) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.Push(cmd, params...))
}

func (mod *IpfsModule) GetBlock(hash string) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.GetBlock(hash))
}

func (mod *IpfsModule) GetFile(hash string) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.GetFile(hash))
}

// func (mod *IpfsModule) GetStream(hash string) (chan []byte, error) {
//	return modules.JsReturnVal(mod.ipfs.GetStream(hash))
// }

func (mod *IpfsModule) GetTree(hash string, depth int) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.GetTree(hash, depth))
}

func (mod *IpfsModule) PushBlock(block []byte) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.PushBlock(block))
}

func (mod *IpfsModule) PushBlockString(block string) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.PushBlockString(block))
}

func (mod *IpfsModule) PushFile(fpath string) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.PushFile(fpath))
}

func (mod *IpfsModule) PushTree(fpath string, depth int) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.PushTree(fpath, depth))
}

// IpfsModule should satisfy KeyManager

func (mod *IpfsModule) ActiveAddress() modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.ActiveAddress(), nil)
}

func (mod *IpfsModule) Addresses() modules.JsObject {
	count := mod.ipfs.AddressCount()
	addresses := make(modules.JsObject)
	array := make([]string, count)

	for i := 0; i < count; i++ {
		addr, _ := mod.ipfs.Address(i)
		array[i] = addr
	}
	addresses["Addresses"] = array
	return modules.JsReturnVal(addresses, nil)
}

func (mod *IpfsModule) Address(n int) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.Address(n))
}

func (mod *IpfsModule) SetAddress(addr string) modules.JsObject {
	err := mod.ipfs.SetAddress(addr)
	if err != nil {
		return modules.JsReturnValErr(err)
	} else {
		// No error means success.
		return modules.JsReturnValNoErr(nil)
	}
}

func (mod *IpfsModule) SetAddressN(n int) modules.JsObject {
	return modules.JsReturnVal(nil, mod.ipfs.SetAddressN(n))
}

func (mod *IpfsModule) NewAddress(set bool) modules.JsObject {
	return modules.JsReturnVal(mod.ipfs.NewAddress(set))
}

func (mod *IpfsModule) AddressCount() modules.JsObject {
	return modules.JsReturnValNoErr(mod.ipfs.AddressCount())
}

// ethereum stores hashes as 32 bytes, but ipfs expects base58 encoding
// thus our convention is that params can be a path, but it must have only a single leading hash (hex encoded)
//  and it must lead with it
// TODO: purpose this...
func (ipfs *Ipfs) Get(cmd string, params ...string) (interface{}, error) {
	// ipfs
	/*
	   n := ipfs.node
	   if len(params) == 0{
	       return ipfs.getCmd(cmd)
	   }*/
	return nil, errors.New("Invalid commmand")
}

func (ipfs *Ipfs) GetObject(hash string) (interface{}, error) {
	// return raw file bytes or a dir tree
	fpath, err := hexPath2B58(hash)
	if err != nil {
		return nil, err
	}
	nd, err := ipfs.node.Resolver.ResolvePath(fpath)
	if err != nil {
		return nil, err
	}

	pb := new(ftpb.Data)
	err = proto.Unmarshal(nd.Data, pb)
	if err != nil {
		return nil, err
	}

	if pb.GetType() == ftpb.Data_Directory {
		return ipfs.GetTree(hash, -1)
	} else {
		return ipfs.GetFile(hash)
	}
}

func (ipfs *Ipfs) GetBlock(hash string) ([]byte, error) {
	h, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	k := util.Key(h)
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)
	b, err := ipfs.node.Blocks.GetBlock(ctx, k)
	if err != nil {
		return nil, fmt.Errorf("block get: %v", err)
	}
	return b.Data, nil
}

func (ipfs *Ipfs) GetFile(hash string) ([]byte, error) {
	h, err := hexPath2B58(hash)
	if err != nil {
		return nil, err
	}
	// buf := bytes.NewBuffer(nil)
	b, err := cat(ipfs.node, []string{h}) //cmds.Cat(ipfs.node, []string{h}, nil, buf)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (ipfs *Ipfs) GetStream(hash string) (chan []byte, error) {
	fpath, err := hexPath2B58(hash)
	if err != nil {
		return nil, err
	}
	dagnode, err := ipfs.node.Resolver.ResolvePath(fpath)
	if err != nil {
		return nil, fmt.Errorf("catFile error: %v", err)
	}
	read, err := uio.NewDagReader(dagnode, ipfs.node.DAG)
	if err != nil {
		return nil, fmt.Errorf("cat error: %v", err)
	}
	ch := make(chan []byte)
	var n int
	go func() {
		for err != io.EOF {
			b := make([]byte, StreamSize)
			// read from reader 1024 bytes at a time
			n, err = read.Read(b)
			if err != nil && err != io.EOF {
				//return nil, err
				break
				// how to handle these errors?!
			}
			// broadcast on channel
			ch <- b[:n]
		}
		close(ch)
	}()
	return ch, nil
}

// TODO: depth
func (ipfs *Ipfs) GetTree(hash string, depth int) (modules.JsObject, error) {
	fpath, err := hexPath2B58(hash)
	if err != nil {
		return nil, err
	}
	nd, err1 := ipfs.node.Resolver.ResolvePath(fpath)
	if err1 != nil {
		return nil, err1
	}
	mhash, err2 := nd.Multihash()
	if err2 != nil {
		return nil, err2
	}
	tree := getTreeNode("", hex.EncodeToString(mhash))
	err3 := grabRefs(ipfs.node, nd, tree)
	return tree, err3
}

func (ipfs *Ipfs) getCmd(cmd string) (interface{}, error) {
	return nil, nil
}

// ...
func (ipfs *Ipfs) Push(cmd string, params ...string) (string, error) {
	return "", errors.New("Invalid cmd")
}

func (ipfs *Ipfs) PushBlock(data []byte) (string, error) {
	b := blocks.NewBlock(data)

	k, err := ipfs.node.Blocks.AddBlock(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString([]byte(k)), nil
}

func (ipfs *Ipfs) PushBlockString(data string) (string, error) {
	return ipfs.PushBlock([]byte(data))
}

func (ipfs *Ipfs) PushFile(fpath string) (string, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	f := &commands.ReaderFile{
		Filename: fpath,
		Reader:   file,
	}
	added := &cmds.AddOutput{}
	nd, err := addFile(ipfs.node, f, added)
	if err != nil {
		return "", err
	}
	h, err := nd.Multihash()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h), nil
	//return ipfs.PushTree(fpath, 1)
}

func (ipfs *Ipfs) PushTree(fpath string, depth int) (string, error) {
	ff, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	f, err := openPath(ff, fpath)
	if err != nil {
		return "", err
	}

	added := &cmds.AddOutput{}
	nd, err := addDir(ipfs.node, f, added)
	if err != nil {
		return "", err
	}
	h, err := nd.Multihash()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h), nil
}

func (ipfs *Ipfs) Subscribe(name string, event string, target string) chan events.Event {
	return nil
}

func (ipfs *Ipfs) UnSubscribe(name string) {
}

// Key manager functions.
// Note in ipfs (in contrast with a blockchain), one is much less likely
// to change keys, as there are accrued benefits to sticking with a single key,
// and there is no notion of "transactions"

// An ipfs ID is simply the multihash of the publickey
func (ipfs *Ipfs) ActiveAddress() string {
	return hex.EncodeToString(ipfs.node.Identity.ID())
}

// Ipfs node's only have one address
func (ipfs *Ipfs) Address(n int) (string, error) {
	return ipfs.ActiveAddress(), nil
}

func (ipfs *Ipfs) SetAddress(addr string) error {
	return fmt.Errorf("It is not possible to set the ipfs node address without restarting.")
}

func (ipfs *Ipfs) SetAddressN(n int) error {
	return fmt.Errorf("It is not possible to set the ipfs node address without restarting.")
}

// We don't create new addresses on the fly
func (ipfs *Ipfs) NewAddress(set bool) (string, error) {
	return "", fmt.Errorf("It is not possible to create new addresses during runtime.")
}

// we only have one ipfs address
func (ipfs *Ipfs) AddressCount() int {
	return 1
}

func HexToB58(s string) (string, error) {
	var b []byte
	if len(s) > 2 {
		if s[:2] == "0x" {
			s = s[2:]
		}
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	bmh := mh.Multihash(b)
	return bmh.B58String(), nil //b58.Encode(b), nil
}

// should this return 0x prefixed?
func B58ToHex(s string) (string, error) {
	r, err := mh.FromB58String(s) //b58.Decode(s) with panic recovery
	if err != nil {
		return "", err
	}
	h := hex.EncodeToString(r)
	return "0x" + h, nil
}

// convert path beginning with 32 byte hex string to path beginning with base58 encoded
func hexPath2B58(p string) (string, error) {
	var err error
	p = strings.TrimLeft(p, "/") // trim leading slash
	spl := strings.Split(p, "/") // split path
	leadingHash := spl[0]
	spl[0], err = HexToB58(leadingHash) // convert leading hash to base58
	if err != nil {
		return "", err
	}

	if len(spl) > 1 {
		return strings.Join(spl, "/"), nil
	}
	return spl[0], nil
}

func getTreeNode(name, hash string) modules.JsObject {
	obj := make(modules.JsObject)
	obj["Nodes"] = make([]modules.JsObject, 0)
	obj["Name"] = name
	obj["Hash"] = hash
	return obj
}

func grabRefs(n *core.IpfsNode, nd *mdag.Node, tree modules.JsObject) error {
	for _, link := range nd.Links {
		h := link.Hash
		newNode := getTreeNode(link.Name, h.B58String())
		nd, err := n.DAG.Get(util.Key(h))
		if err != nil {
			//log.Errorf("error: cannot retrieve %s (%s)", h.B58String(), err)
			return err
		}
		err = grabRefs(n, nd, newNode)
		if err != nil {
			return err
		}
		nds := tree["Nodes"].([]modules.JsObject)
		nds = append(nds, newNode)
	}
	return nil
}
