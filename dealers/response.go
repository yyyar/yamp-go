//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package dealers

import (
	"github.com/satori/go.uuid"
	"github.com/yyyar/yamp-go/api"
	"github.com/yyyar/yamp-go/format"
	"github.com/yyyar/yamp-go/parser"
	"log"
	"sync"
)

//
// ResponseDealer
//
type ResponseDealer struct {
	sync.RWMutex

	bodyFormat format.BodyFormat
	In         chan parser.Response
	handlers   map[uuid.UUID]api.ResponseHandler
}

//
// NewResponseDealer
//
func NewResponseDealer(bodyFormat format.BodyFormat) *ResponseDealer {

	p := &ResponseDealer{
		bodyFormat: bodyFormat,
		In:         make(chan parser.Response),
		handlers:   make(map[uuid.UUID]api.ResponseHandler),
	}

	go p.Loop()
	return p
}

//
// OnResponse
//
func (p *ResponseDealer) OnResponse(uid uuid.UUID, handler api.ResponseHandler) error {

	p.Lock()
	defer p.Unlock()

	p.handlers[uid] = handler

	return nil

}

//
// Loop
//
func (p *ResponseDealer) Loop() {

	for {

		response, ok := <-p.In
		if !ok {
			return
		}

		p.Lock()

		handler, ok := p.handlers[response.RequestUid]

		if !ok {
			log.Println("No handlers for response uri ", response.RequestUid)
			return
		}

		delete(p.handlers, response.RequestUid)

		p.Unlock()

		go handler(&api.Response{p.bodyFormat, nil, nil, &response})
	}
}
