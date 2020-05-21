package sgo

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/hashicorp/golang-lru"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	CacheSize int
	WaitTime  time.Duration
	Compress  string
}

type render struct {
	Config
	cache *lru.Cache
	mutex sync.Mutex
}

type Response struct {
	Status      int
	Content     string
	Header      http.Header
	ContentType string
}

type cachedResponse struct {
	Status            int
	CompressedContent []byte
	Header            http.Header
	ContentType       string
}

func (s *render) urlHash(url string) string {
	h := sha1.New()
	h.Write([]byte(url))
	return fmt.Sprintf("%x\n", h.Sum(nil))
}
func (s *render) getFromCache(url string) (file *Response, ok bool) {
	urlHash := s.urlHash(url)

	if cacheResult, ok := s.cache.Get(urlHash); ok {
		cachedResponse := cacheResult.(cachedResponse)
		in := *bytes.NewBuffer(cachedResponse.CompressedContent)
		var out bytes.Buffer
		r, _ := gzip.NewReader(&in)
		_, _ = io.Copy(&out, r)

		return &Response{
			Status:      cachedResponse.Status,
			Header:      cachedResponse.Header,
			Content:     out.String(),
			ContentType: cachedResponse.ContentType,
		}, true
	}

	return nil, false
}

func (s *render) setCache(url string, response *Response) {
	urlHash := s.urlHash(url)

	var in bytes.Buffer
	b := []byte(response.Content)
	w := gzip.NewWriter(&in)
	_, _ = w.Write(b)
	_ = w.Close()

	s.cache.Add(urlHash, cachedResponse{
		Status:            response.Status,
		CompressedContent: in.Bytes(),
		ContentType:       response.ContentType,
	})
}

func (s *render) getSSR(url string) (response *Response, hitCache bool, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	response = nil
	hitCache = false
	if file, ok := s.getFromCache(url); ok {
		return file, true, nil
	}

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	requestUrl := url

	resp, err := http.Head(requestUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	// run task list
	strChn := make(chan string)
	go func() {
		var res string
		err := chromedp.Run(ctx,
			chromedp.Navigate(requestUrl),
			chromedp.Sleep(s.WaitTime*time.Second),
			chromedp.OuterHTML(`html`, &res, chromedp.NodeVisible, chromedp.ByQuery),
		)
		if err != nil {
			log.Fatal(err)
		}

		strChn <- res
		close(strChn)
	}()

	response = &Response{
		Header:      resp.Header,
		Status:      resp.StatusCode,
		Content:     <-strChn,
		ContentType: contentType,
	}

	s.setCache(url, response)

	return response, false, nil
}

func New(config Config) *render {
	lruCache, _ := lru.New(config.CacheSize)
	return &render{
		Config: config,
		cache:  lruCache,
	}
}
