//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package dealers

import (
	"github.com/yyyar/yamp-go/api"
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
	In         chan parser.Event
	handlers   map[string][]api.EventHandler
}

//
// NewEventDealer
//
func NewEventDealer(bodyFormat format.BodyFormat) *EventDealer {

	e := &EventDealer{
		bodyFormat: bodyFormat,
		In:         make(chan parser.Event),
		handlers:   make(map[string][]api.EventHandler),
	}

	go e.Loop()
	return e
}

//
// OnEvent
//
func (e *EventDealer) OnEvent(uri string, handler api.EventHandler) {

	e.Lock()
	defer e.Unlock()

	if _, ok := e.handlers[uri]; !ok {
		e.handlers[uri] = []api.EventHandler{}
	}

	e.handlers[uri] = append(e.handlers[uri], handler)
}

//
// Loop
//
func (e *EventDealer) Loop() {

	for {

		event, ok := <-e.In

		if !ok {
			return
		}

		e.RLock()

		handlers, ok := e.handlers[event.Uri]

		if !ok {
			log.Println("No handlers for event uri " + event.Uri)
			continue
		}

		for _, handler := range handlers {
			go handler(&api.Event{
				e.bodyFormat,
				event,
			})
		}

		e.RUnlock()
	}
}
