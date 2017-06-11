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
