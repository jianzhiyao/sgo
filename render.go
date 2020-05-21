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

type File struct {
	Content     string
	ContentType string
}

type cachedFile struct {
	CompressedContent []byte
	ContentType       string
}

func (s *render) urlHash(url string) string {
	h := sha1.New()
	h.Write([]byte(url))
	return fmt.Sprintf("%x\n", h.Sum(nil))
}
func (s *render) getFromCache(url string) (file *File, ok bool) {
	urlHash := s.urlHash(url)

	if cacheResult, ok := s.cache.Get(urlHash); ok {
		cachedFile := cacheResult.(cachedFile)
		in := *bytes.NewBuffer(cachedFile.CompressedContent)
		var out bytes.Buffer
		r, _ := gzip.NewReader(&in)
		_, _ = io.Copy(&out, r)

		return &File{
			Content:     out.String(),
			ContentType: cachedFile.ContentType,
		}, true
	}

	return nil, false
}

func (s *render) setCache(url string, file *File) {
	urlHash := s.urlHash(url)

	var in bytes.Buffer
	b := []byte(file.Content)
	w := gzip.NewWriter(&in)
	_, _ = w.Write(b)
	_ = w.Close()

	s.cache.Add(urlHash, cachedFile{
		CompressedContent: in.Bytes(),
		ContentType:       file.ContentType,
	})
}

func (s *render) RenderPageDynamically(relativeUrl string) (file *File, hitCache bool, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	file = nil
	hitCache = false
	if file, ok := s.getFromCache(relativeUrl); ok {
		return file, true, nil
	}

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	requestUrl := s.BackendServer + "/" + relativeUrl

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

	file = &File{
		Content:     <-strChn,
		ContentType: contentType,
	}

	s.setCache(relativeUrl, file)

	return file, false, nil
}

func New(config Config) *render {
	lruCache, _ := lru.New(config.CacheSize)
	return &render{
		Config: config,
		cache:  lruCache,
	}
}
