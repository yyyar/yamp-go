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
// EventHandler
//
type EventHandler func(*Event)

//
// Event represents pub/sub event
//
type Event struct {

	// body serialization format
	BodyFormat format.BodyFormat

	// event frame
	Frame parser.Event
}

//
// Id returns unique event identifier
//
func (e *Event) Id() string {
	return uuid.UUID(e.Frame.Uid).String()
}

//
// ReadTo reads (parses) event body to object
//
func (e *Event) ReadTo(to interface{}) {
	e.BodyFormat.Parse(e.Frame.Body, to)
}

//
// RawBody returns event data as a raw (unparsed) byte array
//
func (e *Event) RawBody() []byte {
	return e.Frame.Body
}
