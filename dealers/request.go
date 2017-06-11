//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package dealers

import (
	"errors"
	"github.com/yyyar/yamp-go/api"
	"github.com/yyyar/yamp-go/format"
	"github.com/yyyar/yamp-go/parser"
	"log"
	"sync"
)

//
// RequestDealer
//
type RequestDealer struct {
	sync.RWMutex

	bodyFormat format.BodyFormat
	In         chan parser.Request
	out        chan parser.Frame
	handlers   map[string]api.RequestHandler
}

//
// NewRequestDealer
//
func NewRequestDealer(bodyFormat format.BodyFormat, out chan parser.Frame) *RequestDealer {

	p := &RequestDealer{
		bodyFormat: bodyFormat,
		In:         make(chan parser.Request),
		out:        out,
		handlers:   make(map[string]api.RequestHandler),
	}

	go p.Loop()
	return p
}

//
// OnRequest
//
func (p *RequestDealer) OnRequest(uri string, handler api.RequestHandler) error {

	p.Lock()
	defer p.Unlock()

	if _, ok := p.handlers[uri]; ok {
		return errors.New("Request handler on uri " + uri + " already exists")
	}

	p.handlers[uri] = handler

	return nil

}

//
// Loop()
//
func (p *RequestDealer) Loop() {

	for {

		request, ok := <-p.In
		if !ok {
			return
		}

		p.RLock()
		handler, ok := p.handlers[request.Uri]
		p.RUnlock()

		if !ok {
			log.Println("No handlers for request uri " + request.Uri)
			return
		}

		go handler(&api.Request{
			p.bodyFormat,
			request,
		}, &api.Response{
			p.bodyFormat,
			p.out,
			&request,
			nil,
		})
	}
}
