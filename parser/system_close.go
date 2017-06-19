//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

const SYSTEM_CLOSE FrameType = 0x01

type CloseCode uint8

const (
	CLOSE_UNKNOWN               CloseCode = 0x00
	CLOSE_VERSION_NOT_SUPPORTED CloseCode = 0x01
	CLOSE_TIMEOUT               CloseCode = 0x02
	CLOSE_REDIRECT              CloseCode = 0x03
)

//
// SystemClose frame
//
type SystemClose struct {
	Code    CloseCode
	Message string
}

func (this *SystemClose) GetType() FrameType {
	return SYSTEM_CLOSE
}

func (this *SystemClose) Parse(buffer io.Reader) error {

	// Code
	if err := utils.Parse(buffer, &this.Code); err != nil {
		return err
	}

	// size of Message
	var size uint16
	if err := utils.Parse(buffer, &size); err != nil {
		return err
	}

	// Message
	message := make([]uint8, size)
	if err := utils.Parse(buffer, &message); err != nil {
		return err
	}
	this.Message = string(message[:])

	return nil
}

func (this *SystemClose) Serialize(writer io.Writer) error {

	utils.Serialize(writer, this.GetType())

	utils.Serialize(writer, this.Code)
	utils.Serialize(writer, uint16(len(this.Message)))
	utils.Serialize(writer, []byte(this.Message))

	return nil
}
