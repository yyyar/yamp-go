//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package format

import (
	"errors"
)

//
// Body formats registry
//
var registry = map[string]BodyFormat{}

//
// Initialize registry
//
func init() {
	registry["json"] = &JsonBodyFormat{}
}

//
// Returns format by name
//
func Get(name string) BodyFormat {
	format := registry[name]
	return format
}

//
// Register new format by name
//
func Register(name string, format BodyFormat) error {

	if _, ok := registry[name]; ok {
		return errors.New("Format with name " + name + " already exists")
	}

	registry[name] = format
	return nil
}

//
// BodyFormat is interface for yamp body format
//
type BodyFormat interface {

	//
	// Serialize object to byte array
	//
	Serialize(interface{}) ([]byte, error)

	//
	// Parse byte array to object
	//
	Parse([]byte, interface{}) error
}
