//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"github.com/satori/go.uuid"
	"github.com/yyyar/yamp-go/format"
	"github.com/yyyar/yamp-go/parser"
)

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
