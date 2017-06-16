//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"errors"
	"github.com/satori/go.uuid"
	"github.com/yyyar/yamp-go/api"
	"github.com/yyyar/yamp-go/dealers"
	"github.com/yyyar/yamp-go/format"
	"github.com/yyyar/yamp-go/parser"
	"github.com/yyyar/yamp-go/transport"
	"log"
)

const (
	YAMP_VERSION = 0x01
)

// Connection is Yamp connection abstraction supports
// events sending/handling and request/response
// processing
type Connection struct {
	isClient bool

	// Transport connection adapter
	conn transport.Connection

	// Yamp protocol parser
	parser *parser.Parser

	// User frames body format parser/serializer
	bodyFormat format.BodyFormat

	// Channel for pushing frames that will be written to other party
	framesOut chan (parser.Frame)

	*dealers.EventDealer
	*dealers.RequestDealer
	*dealers.ResponseDealer
}

//
// NewConnection Creates new instance of Connection wrapping
// transport.Connection and immediately starting read/write loop
//
func NewConnection(isClient bool, conn transport.Connection, bodyFormat format.BodyFormat) (*Connection, error) {

	out := make(chan parser.Frame)

	connection := &Connection{

		isClient:   isClient,
		conn:       conn,
		parser:     parser.NewParser(conn),
		bodyFormat: bodyFormat,

		framesOut: out,

		EventDealer:    dealers.NewEventDealer(bodyFormat),
		RequestDealer:  dealers.NewRequestDealer(bodyFormat, out),
		ResponseDealer: dealers.NewResponseDealer(bodyFormat),
	}

	// Try handshake
	if err := connection.handshake(); err != nil {

		return nil, err
	}

	return connection, nil
}

//
// Perform initial system.handshake
//
func (this *Connection) handshake() error {

	if this.isClient {
		if err := this.handshakeClient(); err != nil {
			return err
		}
	} else {
		if err := this.handshakeServer(); err != nil {
			return err
		}
	}

	go this.readLoop()
	go this.writeLoop()

	return nil
}

//
// Initiate client-side handshake with server
//
func (this *Connection) handshakeClient() error {

	// Send system.handshake

	(&parser.SystemHandshake{
		Version: YAMP_VERSION,
	}).Serialize(this.conn)

	// Get response

	frame, ok := <-this.parser.Frames
	if !ok {
		err := <-this.parser.Error
		return err
	}

	// If got system.handshake back, then we're ok
	if frame.GetType() == parser.SYSTEM_HANDSHAKE {
		return nil
	}

	// Something bad happened
	if frame.GetType() == parser.SYSTEM_CLOSE {
		return errors.New(frame.(*parser.SystemClose).Reason)
	}

	return errors.New("Unexpected event")
}

//
// Handle server party handshake
//
func (this *Connection) handshakeServer() error {

	// Wait for client to send system.handshake

	frame, ok := <-this.parser.Frames
	if !ok {
		err := <-this.parser.Error
		return err
	}

	// If client sent something else, close connection

	if frame.GetType() != parser.SYSTEM_HANDSHAKE {
		this.conn.Close()
		return errors.New("Unexpected frame")
	}

	// Check versions, and if we're satisfied
	// respond with the same system.handshake

	handshake := frame.(*parser.SystemHandshake)
	if handshake.Version != YAMP_VERSION {
		this.conn.Close()
		return errors.New("Version not supported")
	}

	frame.Serialize(this.conn)

	return nil
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
				this.EventDealer.In <- *(frame).(*parser.Event)
			case parser.RESPONSE:
				this.ResponseDealer.In <- *(frame).(*parser.Response)
			case parser.REQUEST:
				this.RequestDealer.In <- *(frame).(*parser.Request)
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
func (this *Connection) SendRequest(uri string, body interface{}, handler api.ResponseHandler) {

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
