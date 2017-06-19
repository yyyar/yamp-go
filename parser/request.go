//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

const REQUEST FrameType = 0x11

//
// Request frame
//
type Request struct {
	UserHeader
	UserBody
}

func (this Request) GetType() FrameType {
	return REQUEST
}

func (this *Request) Parse(buffer io.Reader) error {

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

func (this *Request) Serialize(writer io.Writer) error {

	utils.Serialize(writer, this.GetType())

	WriteUserHeader(writer, this.UserHeader)
	WriteUserBody(writer, this.UserBody)

	return nil
}
