package sgo

import (
	"testing"
)

func TestService_RenderPage(t *testing.T) {
	service := New(Config{
		CacheSize: 1000,
		//return html after waiting
		WaitTime:      3,
		BackendServer: "http://bing.com",
	})

	file1, _, _ := service.RenderPageDynamically("/")
	file2, hitCache2, _ := service.RenderPageDynamically("/")

	if file1.Content != file2.Content {
		t.Fatalf("render error")
	}
	if file1.ContentType != file2.ContentType {
		t.Fatalf("render error")
	}
	if !hitCache2 {
		t.Fatalf("cannot hit cache")
	}
}
