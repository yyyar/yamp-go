//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

const RESPONSE FrameType = 0x08

type ResponseType uint8

const (
	RESPONSE_DONE      ResponseType = 0x00
	RESPONSE_ERROR     ResponseType = 0x01
	RESPONSE_PROGRESS  ResponseType = 0x02
	RESPONSE_CANCELLED ResponseType = 0x03
)

//
// Response frame
//
type Response struct {
	UserHeader
	RequestUid [16]byte
	Type       ResponseType
	UserBody
}

func (this Response) GetType() FrameType {
	return RESPONSE
}

func (this *Response) Parse(buffer io.Reader) error {

	// UserHeader
	header, err := ParseUserHeader(buffer)
	if err != nil {
		return err
	}
	this.UserHeader = *header

	// RequestUid
	if err := utils.Parse(buffer, &this.RequestUid); err != nil {
		return err
	}

	// Type
	if err := utils.Parse(buffer, &this.Type); err != nil {
		return err
	}

	// UserBody
	body, err := ParseUserBody(buffer)
	if err != nil {
		return err
	}
	this.UserBody = *body

	return nil
}

func (this *Response) Serialize(writer io.Writer) error {

	utils.Serialize(writer, this.GetType())

	WriteUserHeader(writer, this.UserHeader)
	utils.Serialize(writer, this.RequestUid)
	utils.Serialize(writer, this.Type)
	WriteUserBody(writer, this.UserBody)

	return nil
}
