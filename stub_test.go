package main

import "testing"

func TestHash(t *testing.T) {
	h := hash("<xml><input><name>pallat</name><email>yod.pallat@gmail.com</email></input></xml>")
	t.Error(h)
}
