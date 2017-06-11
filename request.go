//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"errors"
	"github.com/satori/go.uuid"
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
	in         chan parser.Request
	out        chan parser.Frame
	handlers   map[string]RequestHandler
}

//
// NewRequestDealer
//
func NewRequestDealer(bodyFormat format.BodyFormat, out chan parser.Frame) *RequestDealer {

	p := &RequestDealer{
		bodyFormat: bodyFormat,
		in:         make(chan parser.Request),
		out:        out,
		handlers:   make(map[string]RequestHandler),
	}

	go p.Loop()
	return p
}

//
// OnRequest
//
func (p *RequestDealer) OnRequest(uri string, handler RequestHandler) error {

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

		request, ok := <-p.in
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

		go handler(&Request{
			p.bodyFormat,
			request,
		}, &Response{
			p.bodyFormat,
			p.out,
			&request,
			nil,
		})
	}
}

// --------------------------------------------------------------------------

//
// Request represents request that came to request handler
//
type Request struct {

	// body serialization format
	bodyFormat format.BodyFormat

	// request frame
	frame parser.Request
}

//
// Id returns unique identifier of request
//
func (r *Request) Id() string {
	return uuid.UUID(r.frame.Uid).String()
}

//
// ReadTo reads (parses) request data into object
//
func (r *Request) ReadTo(to interface{}) {
	r.bodyFormat.Parse(r.frame.Body, to)
}

//
// RawBody returns raw (unparsed) request body as byte array
//
func (r *Request) RawBody() []byte {
	return r.frame.Body
}

//
// Progressive indicates request of progressive responses
//
func (r *Request) Progessive() bool {
	return r.frame.Progressive
}
