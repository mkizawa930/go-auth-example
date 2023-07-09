package main

import "testing"

func TestRandomString(t *testing.T) {
	b, err := randomString(32)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(string(b))
}
