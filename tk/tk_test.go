package tk

import (
	"testing"
)

func TestGet(t *testing.T) {
	tkk := "443916.547221231"
	expect := "68957.510801"
	tk, err := Get("hello\u00A0world", tkk)
	if err != nil {
		t.Error(err)
	}
	if tk != expect {
		t.Errorf("wrong tk, expect: %q, got: %q", expect, tk)
	}
}
