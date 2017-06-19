//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

const EVENT FrameType = 0x10

//
// Event frame
//
type Event struct {
	UserHeader
	UserBody
}

func (this *Event) GetType() FrameType {
	return EVENT
}

func (this *Event) Parse(buffer io.Reader) error {

	// UserHeader
	header, err := ParseUserHeader(buffer)
	if err != nil {
		return err
	}
	this.UserHeader = *header

	// UserBody
	body, err := ParseUserBody(buffer)
	if err != nil {
		return err
	}
	this.UserBody = *body

	return nil
}

func (this *Event) Serialize(writer io.Writer) error {

	utils.Serialize(writer, this.GetType())

	WriteUserHeader(writer, this.UserHeader)
	WriteUserBody(writer, this.UserBody)

	return nil
}
