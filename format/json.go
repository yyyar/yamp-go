//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package format

import (
	"encoding/json"
)

//
// JsonBodyFormat
//
type JsonBodyFormat struct{}

//
// Serialize JSON
//
func (this *JsonBodyFormat) Serialize(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

//
// Deserialize JSON
//
func (this *JsonBodyFormat) Parse(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
