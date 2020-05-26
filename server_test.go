package sgo

import "testing"

func TestServe(t *testing.T) {
	NewDefaultServer("https://baidu.com",8886).ListenAndServe()
}