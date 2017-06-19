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
// RequestHandler
//
type RequestHandler func(*Request, *Response)

//
// Request represents request that came to request handler
//
type Request struct {

	// body serialization format
	BodyFormat format.BodyFormat

	// request frame
	Frame parser.Request
}

//
// Id returns unique identifier of request
//
func (r *Request) Id() string {
	return uuid.UUID(r.Frame.Uid).String()
}

//
// Read reads (parses) request data into object
//
func (r *Request) Read(to interface{}) {
	r.BodyFormat.Parse(r.Frame.Body, to)
}

//
// RawBody returns raw (unparsed) request body as byte array
//
func (r *Request) RawBody() []byte {
	return r.Frame.Body
}
