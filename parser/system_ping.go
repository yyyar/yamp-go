//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

const SYSTEM_PING FrameType = 0x01

//
// SystemPing frame
//
type SystemPing struct {
	Payload string
}

func (this *SystemPing) GetType() FrameType {
	return SYSTEM_PING
}

func (this *SystemPing) Parse(buffer io.Reader) error {

	// size of Payload
	var size uint8
	if err := utils.Parse(buffer, &size); err != nil {
		return err
	}

	// Payload
	payload := make([]uint8, size)
	if err := utils.Parse(buffer, &payload); err != nil {
		return err
	}
	this.Payload = string(payload[:])

	return nil
}

func (this *SystemPing) Serialize(writer io.Writer) error {

	utils.Serialize(writer, this.GetType())

	utils.Serialize(writer, uint8(len(this.Payload)))
	utils.Serialize(writer, []byte(this.Payload))

	return nil
}
