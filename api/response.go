//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package api

import (
	"github.com/satori/go.uuid"
	"github.com/yyyar/yamp-go/format"
	"github.com/yyyar/yamp-go/parser"
)

//
// ResponseHandler
//
type ResponseHandler func(*Response)

///
// Response represents response for a request
//
type Response struct {

	// Body serialization format
	BodyFormat format.BodyFormat

	// Output channel to push response when ready
	Out chan parser.Frame

	// Optional request frame for this response
	RequestFrame *parser.Request

	// Response frame
	Frame *parser.Response
}

//
// Id returns unique identifier of response
//
func (r *Response) Id() string {
	return uuid.UUID(r.Frame.Uid).String()
}

//
// RequestId returns unique identifier of Request of this resposne
//
func (r *Response) RequestId() string {

	var uuid uuid.UUID
	if r.RequestFrame != nil {
		uuid = r.RequestFrame.Uid
	} else {
		uuid = r.Frame.RequestUid
	}
	return uuid.String()
}

//
// Read reads (parses) response data into object
//
func (r *Response) Read(to interface{}) {
	r.BodyFormat.Parse(r.Frame.Body, to)
}

//
// IsDone indicates that this is successed response
//
func (r *Response) IsDone() bool {
	return r.Frame.Type == parser.RESPONSE_DONE
}

//
// IsError indicates that this is errored response
//
func (r *Response) IsError() bool {
	return r.Frame.Type == parser.RESPONSE_ERROR
}

//
// IsProgress indicates progressive response
//
func (r *Response) IsProgress() bool {
	return r.Frame.Type == parser.RESPONSE_PROGRESS
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
// Progress sends done response to requester party
//
func (r *Response) Progress(obj interface{}) {
	r.send(parser.RESPONSE_PROGRESS, obj)
}

//
// Serialize body and send out for delivery to other party
//
func (r *Response) send(t parser.ResponseType, obj interface{}) {

	b, _ := r.BodyFormat.Serialize(obj)

	response := parser.Response{
		UserHeader: parser.UserHeader{
			Uid: uuid.NewV1(),
			Uri: r.RequestFrame.Uri,
		},
		RequestUid: r.RequestFrame.Uid,
		Type:       t,
		UserBody: parser.UserBody{
			Body: b,
		},
	}

	r.Out <- &response
}
