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
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	CacheSize int
	WaitTime  time.Duration
	CacheTime int64
	Compress  string
}

type render struct {
	Config
	cache      *lru.Cache
	cacheTime  int64
	mutex      sync.Mutex
	cacheMutex sync.Mutex
}

type Response struct {
	Status      int
	Content     string
	ContentType string
}

type cachedResponse struct {
	Status            int
	CompressedContent []byte
	ContentType       string
	CacheTime         int64
}

func (s *render) urlHash(url string) string {
	h := sha1.New()
	h.Write([]byte(url))
	return fmt.Sprintf("%x\n", h.Sum(nil))
}

func (s *render) getFromCache(url string) (file *Response, ok bool) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	urlHash := s.urlHash(url)

	if cacheResult, ok := s.cache.Get(urlHash); ok {
		cachedResponse := cacheResult.(cachedResponse)

		log.Println(cachedResponse.CacheTime)
		log.Println(s.cacheTime)
		log.Println(time.Now().Unix())
		//cache expired
		if s.cacheTime > 0 && (cachedResponse.CacheTime+s.cacheTime) < time.Now().Unix() {
			s.cache.Remove(urlHash)
			return nil, false
		}

		in := *bytes.NewBuffer(cachedResponse.CompressedContent)
		var out bytes.Buffer
		r, _ := gzip.NewReader(&in)
		_, _ = io.Copy(&out, r)

		return &Response{
			Status:      cachedResponse.Status,
			Content:     out.String(),
			ContentType: cachedResponse.ContentType,
		}, true
	}

	return nil, false
}

func (s *render) setCache(url string, response *Response) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

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
		CacheTime:         time.Now().Unix(),
	})
}

func (s *render) GetSSR(url string) (response *Response, hitCache bool, err error) {
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
	//only render `text/html`
	if contentType == `text/html` {
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
	} else {
		go func() {
			resp, err := http.Get(requestUrl)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			strChn <- string(body)
			close(strChn)
		}()
	}

	response = &Response{
		Status:      resp.StatusCode,
		Content:     <-strChn,
		ContentType: contentType,
	}

	s.setCache(url, response)

	return response, false, nil
}

func NewRender(config Config) *render {
	lruCache, _ := lru.New(config.CacheSize)
	return &render{
		Config:    config,
		cache:     lruCache,
		cacheTime: config.CacheTime,
	}
}
