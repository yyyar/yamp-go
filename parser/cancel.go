//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

const CANCEL FrameType = 0x12

//
// Cancel frame
//
type Cancel struct {
	UserHeader
	RequestUid [16]byte
}

func (this Cancel) GetType() FrameType {
	return CANCEL
}

func (this *Cancel) Parse(buffer io.Reader) error {

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

	return nil
}

func (this *Cancel) Serialize(writer io.Writer) error {

	utils.Serialize(writer, this.GetType())

	WriteUserHeader(writer, this.UserHeader)
	utils.Serialize(writer, this.RequestUid)

	return nil
}
