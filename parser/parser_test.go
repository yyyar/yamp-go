//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"io"
	"testing"
)

//
// Test Parser
//
func TestParser(t *testing.T) {

	reader, writer := io.Pipe()
	parser := NewParser(reader)

	go (&Event{
		UserHeader: UserHeader{
			Uid: [16]byte{3, 3, 3},
			Uri: "test",
		},
		UserBody: UserBody{
			Body: []byte{1, 2, 3, 4, 5},
		},
	}).Serialize(writer)

	frame, ok := <-parser.Frames

	if !ok {
		t.Error("Nothing from parser")
	}

	if frame == nil || frame.(*Event) == nil {
		t.Errorf("Bad frame")
	}

	event := frame.(*Event)
	if event.Uri != "test" || event.Uid != [16]byte{3, 3, 3} {
		t.Error("Bad frame content")
	}

	t.Log(frame)

	// Close writer
	writer.Close()

	// Should be no frames more
	frame, ok = <-parser.Frames

	// And only the EOF error
	err := <-parser.Error

	if err != io.EOF {
		t.Error("Not EOF after closed reader")
	}

}
