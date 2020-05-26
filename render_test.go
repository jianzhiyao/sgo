package sgo

import (
	"testing"
)

func TestService_RenderPage(t *testing.T) {
	service := NewRender(Config{
		CacheSize: 1000,
		//return html after waiting
		WaitTime: 3,
	})

	file1, _, _ := service.getSSR("http://bing.com/")
	file2, hitCache2, _ := service.getSSR("http://bing.com/")

	if file1.Content != file2.Content {
		t.Fatalf("Content error")
	}
	if file1.ContentType != file2.ContentType {
		t.Fatalf("ContentType error")
	}
	if !hitCache2 {
		t.Fatalf("cannot hit cache")
	}
}
