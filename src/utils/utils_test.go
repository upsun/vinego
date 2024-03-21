package utils

import (
	"reflect"
	"testing"
)

func TestRemove(t *testing.T) {
	data := []string{"a", "b", "c"}
	exp := []string{"a", "c"}
	Remove(&data, 1, 1)
	if !reflect.DeepEqual(data, exp) {
		t.Errorf("expected %#v, got %#v", exp, data)
	}
}
