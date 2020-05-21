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

func NewServer(backend string, port int) *http.Server {
	r := gin.New()

	rd := New(Config{
		CacheSize: 1000,
		WaitTime:  3,
	})

	r.Use(func(c *gin.Context) {
		backendUrl := backend + c.Request.URL.Path
		fmt.Println(backendUrl)
		if c.Request.Method != http.MethodGet {
			getProxy(backendUrl, c)
			return
		}

		if true {
			response, hitCache, err := rd.getSSR(backendUrl)

			log.Println("request:", c.Request.URL.Path)
			log.Println("hitCache:", hitCache)
			log.Println("err:", err)
			c.String(response.Status, response.Content)
			return
		} else {
			getProxy(backendUrl, c)
			return
		}
	})

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", "0.0.0.0", port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func getProxy(backendUrl string, c *gin.Context) {
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
