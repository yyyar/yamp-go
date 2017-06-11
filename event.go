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
