package events

// Events are passed from modules to the decerver event handler. They should implement the Event
// interface. The event system is pub/sub. If you want an object to subscribe to events, make sure
// it implements the Subscriber interface and pass it to the event system.
import (
	"time"
)

type Event interface {
	// Event ID (for ethereum it could be newBlock or newTx:post)
	Event() string
	// object of the event
	Target() string
	// The event data.
	Resource() interface{}
	// The source is the id of the module that produced the event.
	Source() string
	// Timestamp is written by the module as it's being passed to the event handler.
	Timestamp() time.Time
}

// This interface allow modules to subscribe to and publish events. It is implemented by the 
// event processor.
type EventRegistry interface {
	Post(e Event)
	Subscribe(sub Subscriber)
	Unsubscribe(sub Subscriber)
}

// A default object that implements 'Event'
type DefaultEvent struct {
	event    string
	target   string
	resource interface{}
	source   string
	ts       time.Time
}

// Use this to get a generic event object.
func NewDefaultEvent() *DefaultEvent{
	return &DefaultEvent{}
}

func (e *DefaultEvent) Event() string {
	return e.event
}

func (e *DefaultEvent) Target() string {
	return e.target
}

func (e *DefaultEvent) Resource() interface{} {
	return e.resource
}

func (e *DefaultEvent) Source() string {
	return e.source
}

func (e *DefaultEvent) Timestamp() time.Time {
	return e.ts
}

func (e *DefaultEvent) GetCurrentTime() {
	e.ts = time.Now()
}

type Subscriber interface {
	Channel() chan Event
	Source() string
}