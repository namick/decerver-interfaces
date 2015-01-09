package ipfs

import (
	//"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
	//"path"
	"github.com/eris-ltd/decerver-interfaces/modules"
	"github.com/eris-ltd/decerver-interfaces/glue/ipfs"
	"testing"
)

var (

	IPFS ipfs.IpfsModule

	block = `here is a block of data to push. it is a modest size amount.
    not too much data, but not too little.
    Really, just the right amount for a 1, 2, puncharoo.
    Or would you prefer I went on about shakespeare?
    Alas, I have neither monkeys nor typewriters.
    This will have to do`

	block2 = `this block has much less data. don't mock him`

	blockMap = map[string]string{
		"block.txt":  block,
		"block2.txt": block2,
	}

	tree = modules.FsNode{
		Nodes: []*modules.FsNode{
			&modules.FsNode{
				Nodes: []*modules.FsNode{
					&modules.FsNode{
						Name: "block.txt",
					},
				},
				Name: "bar",
			},
			&modules.FsNode{
				Nodes: []*modules.FsNode{
					&modules.FsNode{
						Name: "block2.txt",
					},
				},
				Name: "baz",
			},
		},
		Name: "mytree",
	}
)

func init(){
	IPFS = ipfs.NewIpfsModule()
}

func writeFile(t *testing.T, name string, data []byte) {
	err := ioutil.WriteFile(name, []byte(block), 0600)
	if err != nil {
		t.Fatal(err)
	}
}

func rmFile(t *testing.T, name string) {
	err := os.Remove(name)
	if err != nil {
		t.Fatal(err)
	}
}

func mkTree(t *testing.T, thisTree *modules.FsNode, dir string) {
	dir += "/" + thisTree.Name
	if len(thisTree.Nodes) == 0 {
		writeFile(t, dir, []byte(blockMap[thisTree.Name]))
	} else {
		err := os.Mkdir(dir, 0777)
		if err != nil {
			t.Fatal(err)
		}
		for _, tr := range thisTree.Nodes {
			mkTree(t, tr, dir)
		}
	}
}

func rmTree(t *testing.T, name string) {
	err := os.RemoveAll(name)
	if err != nil {
		t.Fatal(err)
	}
}

// we can't just compare hashes sincethe tree we construct
// above has no hashes...
func cmpTree(t *testing.T, tree1 *modules.FsNode, tree2 *modules.FsNode) {
	if tree1.Name != tree2.Name {
		t.Fatal("trees have different names")
	}
	if len(tree1.Nodes) != len(tree2.Nodes) {
		t.Fatal("trees have different link lengths")
	}
	for i, _ := range tree1.Nodes {
		cmpTree(t, tree1.Nodes[i], tree2.Nodes[i])
	}
}

func TestModule(t *testing.T) {
	
	// test IpfsModule satisfies DecerverModule
	f := func(b modules.Module) {}
	f(IPFS)

}


func TestBlock(t *testing.T) {
	h, err := IPFS.PushBlock([]byte(block))
	if err != nil {
		t.Error(err)
	}
	b, err := IPFS.GetBlock(h)
	if err != nil {
		t.Error(err)
	}
	if string(b) != block {
		t.Error("Expected: %s, Got: %s", block, string(b))
	}
}

func TestFile(t *testing.T) {
	filename := ".test"
	writeFile(t, filename, []byte(block))
	defer rmFile(t, filename)
	h, err := IPFS.PushFile(filename)
	if err != nil {
		t.Error(err)
	}
	b, err := IPFS.GetFile(h)
	if err != nil {
		t.Error(err)
	}
	if string(b) != block {
		t.Error("Expected: %s, Got: %s", block, string(b))
	}
}

func TestStream(t *testing.T) {
	StreamSize = 32
	filename := ".test"
	writeFile(t, filename, []byte(block))
	defer rmFile(t, filename)
	h, err := IPFS.PushFile(filename)
	if err != nil {
		t.Error(err)
	}
	ch, err := IPFS.GetStream(h)
	if err != nil {
		t.Fatal(err.Error())
	}
	b := ""
	for r := range ch {
		b += string(r)
	}
	if string(b) != block {
		t.Error("Expected: %s, Got: %s", block, string(b))
	}
}

func TestTree(t *testing.T) {
	mkTree(t, &tree, ".")
	defer rmTree(t, tree.Name)
	h, err := IPFS.PushTree(tree.Name, -1)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := IPFS.GetTree(h, -1)
	if err != nil {
		t.Fatal(err)
	}
	tr.Name = tree.Name
	cmpTree(t, tr, &tree)
}

func TestShutdown(t *testing.T) {
	IPFS.Shutdown()
	time.Sleep(time.Second * 5)
	IPFS.ipfs.node = nil
	log.Println("restarting...")
	IPFS = start(true)
	time.Sleep(time.Second * 5)
}
*/