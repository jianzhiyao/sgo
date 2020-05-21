package sgo

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

func NewServer(backend string) *http.Server {
	r := gin.New()

	rd := New(Config{
		CacheSize: 1000,
		WaitTime:  3,
	})

	r.Use(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			getProxy(c)
			return
		}

		if true {
			response, hitCache, err := rd.getSSR(backend + c.Request.URL.Path)

			log.Println("request:", c.Request.URL.Path)
			log.Println("hitCache:", hitCache)
			log.Println("err:", err)
			c.String(response.Status, response.Content)
			return
		} else {
			getProxy(c)
			return
		}
	})

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", "0.0.0.0", 8887),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func getProxy(c *gin.Context) {
	director := func(req *http.Request) {
		req.Host = c.Request.Host
		req.URL.Scheme = c.Request.URL.Scheme
		req.URL.Host = c.Request.URL.Host
		req.RequestURI = c.Request.RequestURI
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
}
