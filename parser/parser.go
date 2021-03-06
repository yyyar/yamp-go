//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

//
// Frames factory
//
var framesFactory = map[FrameType](func() Frame){}

//
// Initialize module: add frames factory functions
//
func init() {
	framesFactory[SYSTEM_HANDSHAKE] = (func() Frame { return &SystemHandshake{} })
	framesFactory[SYSTEM_PING] = (func() Frame { return &SystemPing{} })
	framesFactory[SYSTEM_CLOSE] = (func() Frame { return &SystemClose{} })

	framesFactory[EVENT] = (func() Frame { return &Event{} })
	framesFactory[REQUEST] = (func() Frame { return &Request{} })
	framesFactory[CANCEL] = (func() Frame { return &Cancel{} })
	framesFactory[RESPONSE] = (func() Frame { return &Response{} })
}

//
// Frame type represents specific frame
//
type FrameType uint8

//
// Frame is transferred frame with concrete type
//
type Frame interface {

	// Returns frame type
	GetType() FrameType

	// Parse itself from reader
	Parse(reader io.Reader) error

	// Write itself to writer
	Serialize(writer io.Writer) error
}

//
// Frames Parser
//
type Parser struct {

	// Reader to read bytes to parse from
	reader io.Reader

	// Channel to push parsed frames
	Frames chan Frame

	// Error to push reason of parser stop
	Error chan error
}

//
// Creates new instance of Parser and starts parsing loop
//
func NewParser(reader io.Reader) *Parser {

	parser := Parser{
		reader: reader,
		Frames: make(chan Frame),
		Error:  make(chan error, 1),
	}

	go parser.loop()

	return &parser
}

//
// Parsing loop
//
func (this *Parser) loop() {
	for {
		frame, err := this.nextFrame()

		if err != nil {
			close(this.Frames)
			this.Error <- err
			return
		}

		this.Frames <- frame
	}
}

//
// Parse next frame
//
func (this *Parser) nextFrame() (Frame, error) {

	var frameType FrameType
	if err := utils.Parse(this.reader, &frameType); err != nil {
		return nil, err
	}

	frame := framesFactory[frameType]()
	if err := frame.Parse(this.reader); err != nil {
		return nil, err
	}

	return frame, nil
}
