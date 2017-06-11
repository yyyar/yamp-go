//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"github.com/satori/go.uuid"
	"github.com/yyyar/yamp-go/format"
	"github.com/yyyar/yamp-go/parser"
	"github.com/yyyar/yamp-go/transport"
	"log"
)

// EventHandler
type EventHandler func(*Event)

// RequestHandler
type RequestHandler func(*Request, *Response)

// ResponseHandler
type ResponseHandler func(*Response)

// Connection is Yamp connection abstraction supports
// events sending/handling and request/response
// processing
type Connection struct {

	// Transport connection adapter
	conn transport.Connection

	// Yamp protocol parser
	parser *parser.Parser

	// User frames body format parser/serializer
	bodyFormat format.BodyFormat

	// Channel for pushing frames that will be written to other party
	framesOut chan (parser.Frame)

	*EventDealer
	*RequestDealer
	*ResponseDealer
}

//
// NewConnection Creates new instance of Connection wrapping
// transport.Connection and immediately starting read/write loop
//
func NewConnection(conn transport.Connection, bodyFormat format.BodyFormat) *Connection {

	out := make(chan parser.Frame)

	connection := &Connection{

		conn:       conn,
		parser:     parser.NewParser(conn),
		bodyFormat: bodyFormat,

		framesOut: out,

		EventDealer:    NewEventDealer(bodyFormat),
		RequestDealer:  NewRequestDealer(bodyFormat, out),
		ResponseDealer: NewResponseDealer(bodyFormat),
	}

	go connection.readLoop()
	go connection.writeLoop()

	return connection
}

//
// Serializing loop
//
func (this *Connection) writeLoop() {

	for {
		frame, ok := <-this.framesOut

		if !ok {
			return
		}

		frame.Serialize(this.conn)
	}

}

//
// Parsing loop
//
func (this *Connection) readLoop() {

	for {
		select {

		// Got new parsed frame
		case frame, ok := <-this.parser.Frames:

			if !ok {
				log.Println(<-this.parser.Error)
				return
			}

			// Dispatch new frame
			switch frame.GetType() {
			case parser.EVENT:
				this.EventDealer.in <- *(frame).(*parser.Event)
			case parser.RESPONSE:
				this.ResponseDealer.in <- *(frame).(*parser.Response)
			case parser.REQUEST:
				this.RequestDealer.in <- *(frame).(*parser.Request)
			default:
				log.Println("Unhandled frame", frame.GetType(), frame)

			}
		}

	}

}

//
// SendEvent
//
func (this *Connection) SendEvent(uri string, body interface{}) {

	uid := uuid.NewV1()
	b, _ := this.bodyFormat.Serialize(body)

	event := parser.Event{
		UserHeader: parser.UserHeader{
			Uid: uid,
			Uri: uri,
		},
		UserBody: parser.UserBody{
			Body: b,
		},
	}

	this.framesOut <- &event

}

//
// SendRequest
//
func (this *Connection) SendRequest(uri string, body interface{}, handler ResponseHandler) {

	uid := uuid.NewV1()
	b, _ := this.bodyFormat.Serialize(body)

	request := parser.Request{
		UserHeader: parser.UserHeader{
			Uid: uid,
			Uri: uri,
		},
		Progressive: false,
		UserBody: parser.UserBody{
			Body: b,
		},
	}

	this.OnResponse(uid, handler)

	this.framesOut <- &request
}

//
// Drop connection instantly
//
func (this *Connection) Destroy() {
	this.conn.Close()
}
