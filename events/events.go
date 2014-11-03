package events

// Events are passed from modules to the decerver event handler. They should implement the Event
// interface. The event system is pub/sub. If you want an object to subscribe to events, make sure
// it implements the Subscriber interface and pass it to the event system.

import (
	"time"
)

type Event struct {
	// Event ID (for ethereum it could be newBlock or newTx:post)
	Event string
    // object of the event
    Target string
	// The event data.
	Resource interface{}
	// The source is the id of the module that produced the event.
	Source string
	// Timestamp is written by the module as it's being passed to the event handler.
	Timestamp *time.Time
}

type EventSystem interface {
	Post(e Event)
	AddListener()
}

// A default object that implements 'Event'
type DefaultEvent struct {
	id string
	data interface{}
	source string
	ts uint64
}

func (e *DefaultEvent) Id() string {
	return e.id
}

func (e *DefaultEvent) Data() interface{} {
	return e.data
}

func (e *DefaultEvent) Source() string {
	return e.source
}

func (e *DefaultEvent) Timestamp() uint64 {
	return e.ts
}

type Subscriber interface {
	Post(e Event)
	Filter() []string
}
