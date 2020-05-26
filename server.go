package sgo

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func NewDefaultServer(backend string, port , WaitSecond int) *http.Server {
	r := gin.New()

	rd := NewRender(Config{
		CacheSize: 1000,
		WaitTime:  time.Duration(WaitSecond),
	})

	r.Use(func(c *gin.Context) {
		backendUrl := backend + c.Request.URL.Path
		if c.Request.Method != http.MethodGet {
			GetProxy(backendUrl, c)
			return
		}

		response, hitCache, err := rd.GetSSR(backendUrl)

		log.Println("request:", c.Request.URL.Path)
		log.Println("hitCache:", hitCache)
		log.Println("ContentType:", response.ContentType)
		log.Println("err:", err)
		c.Header(`Content-Type`, response.ContentType)
		c.String(response.Status, response.Content)
		return

	})

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", "0.0.0.0", port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func GetProxy(backendUrl string, c *gin.Context) {
	back, _ := url.Parse(backendUrl)
	director := func(req *http.Request) {
		req.Host = back.Host
		req.URL.Scheme = back.Scheme
		req.URL.Host = back.Host
		req.RequestURI = c.Request.RequestURI
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
}
