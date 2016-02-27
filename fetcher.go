// Copyright (c) 2016 Sebastian Ohm.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// fetcher is a simple key/value store.
type fetcher interface {
	fetch(string) ([]byte, error)
}

// httpFetcher implements a fetcher using a http.Client.
type httpFetcher struct {
	url *url.URL
}

func newHTTPFetcher(baseURL string) (*httpFetcher, error) {
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	fetcher := &httpFetcher{
		url: url,
	}

	return fetcher, nil
}

func (h *httpFetcher) fetch(key string) ([]byte, error) {
	req, err := http.NewRequest("GET", h.url.String()+key, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to GET %s: %d", key, resp.StatusCode)
	}

	val, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return val, nil
}

type cacheItem struct {
	t   time.Time
	val []byte
}

type cacheRequest struct {
	c   chan cacheResponse
	key string
}

type cacheResponse struct {
	err error
	val []byte
}

type cacheFetcher struct {
	sync.RWMutex

	// As it is, this map used as in-memory cache will only consume more memory
	// over time. Given the small, mostly static nature of the cached dataset
	// (bounded by number of pipeline_groups * number of pipelines) that should
	// be ok.
	cache map[string]*cacheItem

	fetcher fetcher
	maxAge  time.Duration
	maxWait time.Duration
	reqs    chan cacheRequest
}

func newCacheFetcher(f fetcher, d, w time.Duration, n int) (*cacheFetcher, error) {
	c := &cacheFetcher{
		cache:   map[string]*cacheItem{},
		reqs:    make(chan cacheRequest),
		fetcher: f,
		maxAge:  d,
		maxWait: w,
	}

	// Delegate 2 concurrent requests to the wrapped fetcher.
	for i := 0; i < n; i++ {
		go c.process()
	}

	return c, nil
}

func (f *cacheFetcher) fetch(k string) ([]byte, error) {
	req := cacheRequest{c: make(chan cacheResponse, 1), key: k}

	f.reqs <- req

	select {
	case res := <-req.c:
		return res.val, res.err
	case <-time.After(f.maxWait):
		return nil, fmt.Errorf("Waited more than %v for %s", f.maxWait, k)
	}
}

func (f *cacheFetcher) get(k string) *cacheItem {
	f.RLock()
	defer f.RUnlock()

	if i, ok := f.cache[k]; ok {
		if time.Since(i.t) <= f.maxAge {
			return i
		}
	}

	return nil
}

func (f *cacheFetcher) set(k string, v []byte) {
	f.Lock()
	defer f.Unlock()

	f.cache[k] = &cacheItem{t: time.Now(), val: v}
}

func (f *cacheFetcher) process() {
	for req := range f.reqs {
		// Fetch the item from the cache.
		if i := f.get(req.key); i != nil {
			req.c <- cacheResponse{err: nil, val: i.val}
			close(req.c)
			continue
		}

		// Fetch the item via the wrapped fetcher.
		v, err := f.fetcher.fetch(req.key)
		if err != nil {
			req.c <- cacheResponse{err: err, val: nil}
			close(req.c)
			continue
		}

		// Cache the fetched item.
		f.set(req.key, v)

		req.c <- cacheResponse{err: nil, val: v}
		close(req.c)
	}
}
