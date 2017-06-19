//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package format

import (
	"reflect"
	"testing"
)

//
// Test json formt
//
func TestJsonFormat(t *testing.T) {

	j := JsonBodyFormat{}

	if j.GetType() != "json" {
		t.Fatal("json type is not json")
	}

	// Test serialize

	obj := map[string][]int{}
	obj["hello"] = []int{1, 2, 3}
	bytes, err := j.Serialize(obj)

	if err != nil {
		t.Fatal(err)
	}

	if string(bytes) != "{\"hello\":[1,2,3]}" {
		t.Fatal("json seerializer got wrong data " + string(bytes))
	}

	// Test parse back

	obj2 := map[string][]int{}
	err = j.Parse(bytes, &obj2)

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(obj2, obj) {
		t.Fatal("Parsed value != serialized value")
	}

}
