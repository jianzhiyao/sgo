package sgo

import "testing"

func TestServe(t *testing.T) {
	NewServer("https://bing.com").ListenAndServe()
}