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
// ResponseDealer
//
type ResponseDealer struct {
	sync.RWMutex

	bodyFormat format.BodyFormat
	in         chan parser.Response
	handlers   map[uuid.UUID]ResponseHandler
}

//
// NewResponseDealer
//
func NewResponseDealer(bodyFormat format.BodyFormat) *ResponseDealer {

	p := &ResponseDealer{
		bodyFormat: bodyFormat,
		in:         make(chan parser.Response),
		handlers:   make(map[uuid.UUID]ResponseHandler),
	}

	go p.Loop()
	return p
}

//
// OnResponse
//
func (p *ResponseDealer) OnResponse(uid uuid.UUID, handler ResponseHandler) error {

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

		response, ok := <-p.in
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

		go handler(&Response{p.bodyFormat, nil, nil, &response})
	}
}

// ---------------------------------------------------------------------

//
// Response represents response for a request
//
type Response struct {

	// Body serialization format
	bodyFormat format.BodyFormat

	// Output channel to push response when ready
	out chan parser.Frame

	// Optional request frame for this response
	requestFrame *parser.Request

	// Response frame
	frame *parser.Response
}

//
// Id returns unique identifier of response
//
func (r *Response) Id() string {
	return uuid.UUID(r.frame.Uid).String()
}

//
// RequestId returns unique identifier of Request of this resposne
//
func (r *Response) RequestId() string {

	var uuid uuid.UUID
	if r.requestFrame != nil {
		uuid = r.requestFrame.Uid
	} else {
		uuid = r.frame.RequestUid
	}
	return uuid.String()
}

//
// ReadTo reads (parses) response data into object
//
func (r *Response) ReadTo(to interface{}) {
	r.bodyFormat.Parse(r.frame.Body, to)
}

//
// IsError indicates that this is errored response
//
func (r *Response) IsError() bool {
	return r.frame.Type == parser.RESPONSE_ERROR
}

//
// IsDone indicates that this is successed response
//
func (r *Response) IsDone() bool {
	return r.frame.Type == parser.RESPONSE_DONE
}

//
// Done sends done response to requester party
//
func (r *Response) Done(obj interface{}) {
	r.send(parser.RESPONSE_DONE, obj)
}

//
// Error sends error response to requester party
//
func (r *Response) Error(obj interface{}) {
	r.send(parser.RESPONSE_ERROR, obj)
}

//
// Serialize body and send out for delivery to other party
//
func (r *Response) send(t parser.ResponseType, obj interface{}) {

	b, _ := r.bodyFormat.Serialize(obj)

	response := parser.Response{
		UserHeader: parser.UserHeader{
			Uid: uuid.NewV1(),
			Uri: r.requestFrame.Uri,
		},
		RequestUid: r.requestFrame.Uid,
		Type:       t,
		UserBody: parser.UserBody{
			Body: b,
		},
	}

	r.out <- &response
}
