package tkk

import "testing"

func TestGet(t *testing.T) {
	tkk, err := Get()
	if err != nil {
		t.Error(err)
	}
	if tkk == "" {
		t.Error("get invalid tkk")
	}
}
