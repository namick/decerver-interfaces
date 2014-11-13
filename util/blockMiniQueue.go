package util

import (
	"container/list"
	"sync"
	"github.com/eris-ltd/decerver-interfaces/modules"
)


// A concurrent queue implementation for *BlockMiniData objects. This is not a
// performance bottleneck, so working with a list is fine.
type BlockMiniQueue struct {
	mutex *sync.Mutex
	queue *list.List
}

func NewBlockMiniQueue() *BlockMiniQueue{
	bmq := &BlockMiniQueue{}
	bmq.queue = list.New()
	bmq.mutex = &sync.Mutex{}
	return bmq
}

func (bmq *BlockMiniQueue) Pop() *modules.BlockMiniData {
	bmq.mutex.Lock()
	val := bmq.queue.Front()
	bmq.queue.Remove(val)
	num, _ := val.Value.(*modules.BlockMiniData)
	bmq.mutex.Unlock()
	return num
}

func (bmq *BlockMiniQueue) Push(bmd *modules.BlockMiniData) {
	bmq.mutex.Lock()
	bmq.queue.PushBack(bmd)
	bmq.mutex.Unlock()
}

func (bmq *BlockMiniQueue) IsEmpty() bool {
	return bmq.queue.Len() == 0
}