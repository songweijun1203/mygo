package engine

import (
	"reflect"
	"testing"
)

func TestClone(t *testing.T) {
	player := &Player{}
	player.sprite = player
	ts := reflect.TypeOf(player.sprite)
	name := ts.String()
	ts2 := ts.Elem()
	name2 := ts2.String()

	var a reflect.Value = reflect.New(ts2)
	var b reflect.Value = a.Elem()

	_ = name
	_ = name2
	_ = b

}
