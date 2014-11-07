package events

// Events are passed from modules to the decerver event handler. They should implement the Event
// interface. The event system is pub/sub. If you want an object to subscribe to events, make sure
// it implements the Subscriber interface and pass it to the event system.
import (
	"time"
	"fmt"
	"github.com/eris-ltd/deCerver-interfaces/core"
)

// This interface allow modules to subscribe to and publish events. It is implemented by the 
// event processor.
type EventRegistry interface {
	Post(e Event)
	Subscribe(sub Subscriber)
	Unsubscribe(sub Subscriber)
}

// A default object that implements 'Event'
type Event struct {
	Event       string
	Target      string
	Resource    interface{}
	Source      string
	TimeStamp   time.Time
}


// A subscriber listens to events.
type Subscriber interface {
	// Events will be passed on this channel
	Channel() chan Event
	// The subscriber only listen to events published by this source
	Source() string
	// The subscriber Id (normally the module or dapp name).
	Id() string
	// This is called when the subscription is removed. Could for example be used to terminate the
	// channel reading process.
	Close()
}

type AteSub struct {
	eventChan chan Event
	closeChan chan bool
	id        string
	source    string
	callback  string
}

func NewAteSub(id, source, callback string, ate core.Runtime) *AteSub {
	as := &AteSub{}
	as.eventChan = make(chan Event)
	as.closeChan = make(chan bool)
	as.source = source
	as.callback = callback
	as.id = id
	
	// Launch the sub channel.
	go func(as *AteSub) {
		fmt.Println("RUNNING ATE EVENT LOOP")
		for {
			select {
			case evt := <- as.eventChan:
				ate.CallFuncOnObj("EventProcessor","Post", evt)
			case <-as.closeChan:
				return
			}
		}
	}(as)
	
	return as
}

func (as *AteSub) Channel() chan Event {
	return as.eventChan
}

func (as *AteSub) Source() string {
	return as.source
}

func (as *AteSub) Id() string {
	return as.id
}

func (as *AteSub) Close() {
	as.closeChan <- true
}

func (as *AteSub) Callback() string {
	return as.callback
}
