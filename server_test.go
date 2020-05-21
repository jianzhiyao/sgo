package sgo

import "testing"

func TestServe(t *testing.T) {
	NewServer("http://baidu.com").ListenAndServe()
}