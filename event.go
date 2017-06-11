//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"github.com/satori/go.uuid"
	"github.com/yyyar/yamp-go/format"
	"github.com/yyyar/yamp-go/parser"
	"log"
	"sync"
)

//
// EventDealer
//
type EventDealer struct {
	sync.RWMutex

	bodyFormat format.BodyFormat
	in         chan parser.Event
	handlers   map[string][]EventHandler
}

//
// NewEventDealer
//
func NewEventDealer(bodyFormat format.BodyFormat) *EventDealer {

	e := &EventDealer{
		bodyFormat: bodyFormat,
		in:         make(chan parser.Event),
		handlers:   make(map[string][]EventHandler),
	}

	go e.Loop()
	return e
}

//
// OnEvent
//
func (e *EventDealer) OnEvent(uri string, handler EventHandler) {

	e.Lock()
	defer e.Unlock()

	if _, ok := e.handlers[uri]; !ok {
		e.handlers[uri] = []EventHandler{}
	}

	e.handlers[uri] = append(e.handlers[uri], handler)
}

//
// Loop
//
func (e *EventDealer) Loop() {

	for {

		event, ok := <-e.in

		if !ok {
			return
		}

		e.RLock()

		handlers, ok := e.handlers[event.Uri]

		if !ok {
			log.Println("No handlers for event uri " + event.Uri)
			return
		}

		for _, handler := range handlers {
			go handler(&Event{
				e.bodyFormat,
				event,
			})
		}

		e.RUnlock()
	}
}

//
// ----------------------------------------------------------------------------------------------
//

//
// Event represents pub/sub event
//
type Event struct {

	// body serialization format
	bodyFormat format.BodyFormat

	// event frame
	frame parser.Event
}

//
// Id returns unique event identifier
//
func (e *Event) Id() string {
	return uuid.UUID(e.frame.Uid).String()
}

//
// ReadTo reads (parses) event body to object
//
func (e *Event) ReadTo(to interface{}) {
	e.bodyFormat.Parse(e.frame.Body, to)
}

//
// RawBody returns event data as a raw (unparsed) byte array
//
func (e *Event) RawBody() []byte {
	return e.frame.Body
}
