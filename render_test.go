package sgo

import (
	"testing"
)

func TestService_RenderPage(t *testing.T) {
	service := New(Config{
		CacheSize: 1000,
		//return html after waiting
		WaitTime:  3,
		BackendServer: "http://bing.com",
	})

	html1,_  :=service.RenderPageDynamically("/")
	html2,hitCache2  :=service.RenderPageDynamically("/")

	if html1 != html2 {
		t.Fatalf("render error")
	}
	if !hitCache2 {
		t.Fatalf("cannot hit cache")
	}
}
