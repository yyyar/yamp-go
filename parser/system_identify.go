//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

const SYSTEM_IDENTIFY FrameType = 0x00

//
// SystemIdentify frame
//
type SystemIdentify struct {
	Version    uint16
	Serializer string
}

func (this *SystemIdentify) GetType() FrameType {
	return SYSTEM_IDENTIFY
}

func (this *SystemIdentify) Parse(buffer io.Reader) error {

	// Version
	if err := utils.Parse(buffer, &this.Version); err != nil {
		return err
	}

	// size of Serializer
	var size uint8
	if err := utils.Parse(buffer, &size); err != nil {
		return err
	}

	// Serializer
	serializer := make([]uint8, size)
	if err := utils.Parse(buffer, &serializer); err != nil {
		return err
	}
	this.Serializer = string(serializer[:])

	return nil
}

func (this *SystemIdentify) Serialize(writer io.Writer) error {

	utils.Serialize(writer, this.GetType())

	utils.Serialize(writer, this.Version)
	utils.Serialize(writer, uint8(len(this.Serializer)))
	utils.Serialize(writer, []byte(this.Serializer))

	return nil
}
