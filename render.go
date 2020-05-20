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
	"sync"
	"time"
)

type Config struct {
	CacheSize     int
	WaitTime      time.Duration
	Compress      string
	BackendServer string
}

type render struct {
	Config
	cache *lru.Cache
	mutex sync.Mutex
}

func (s *render) urlHash(url string) string {
	h := sha1.New()
	h.Write([]byte(url))
	return fmt.Sprintf("%x\n", h.Sum(nil))
}
func (s *render) getFromCache(url string) (html string, ok bool) {
	urlHash := s.urlHash(url)

	if cacheResult, ok := s.cache.Get(urlHash); ok {
		in := *bytes.NewBuffer(cacheResult.([]byte))
		var out bytes.Buffer
		r, _ := gzip.NewReader(&in)
		_, _ = io.Copy(&out, r)

		return out.String(), true
	}

	return "", false
}

func (s *render) setCache(url, html string) {
	urlHash := s.urlHash(url)

	var in bytes.Buffer
	b := []byte(html)
	w := gzip.NewWriter(&in)
	_, _ = w.Write(b)
	_ = w.Close()

	fmt.Println(len(b), len(in.Bytes()))
	s.cache.Add(urlHash, in.Bytes())
}

func (s *render) RenderPageDynamically(relativeUrl string) (html string, hitCache bool) {
	s.mutex.Lock()
	s.mutex.Unlock()

	html = ""
	hitCache = false
	if cacheResult, ok := s.getFromCache(relativeUrl); ok {
		return cacheResult, true
	}

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// run task list
	strChn := make(chan string)
	go func() {
		var res string
		err := chromedp.Run(ctx,
			chromedp.Navigate(s.BackendServer+"/"+relativeUrl),
			chromedp.Sleep(s.WaitTime*time.Second),
			chromedp.OuterHTML(`html`, &res, chromedp.NodeVisible, chromedp.ByQuery),
		)
		if err != nil {
			log.Fatal(err)
		}

		strChn <- res
		close(strChn)
	}()

	html = <-strChn

	s.setCache(relativeUrl, html)

	return html, false
}

func New(config Config) *render {
	lruCache, _ := lru.New(config.CacheSize)
	return &render{
		Config: config,
		cache:  lruCache,
	}
}
