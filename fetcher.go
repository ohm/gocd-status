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
	v []byte
	t time.Time
}

// cacheFetcher implements a simple in-memory read through caching strategy on
// top of another fetcher.
type cacheFetcher struct {
	sync.RWMutex

	fetcher fetcher
	items   map[string]*cacheItem
	maxAge  time.Duration
}

func newCacheFetcher(fetcher fetcher, maxAge time.Duration) (*cacheFetcher, error) {
	f := &cacheFetcher{
		fetcher: fetcher,
		items:   map[string]*cacheItem{},
		maxAge:  maxAge,
	}

	return f, nil
}

func (f *cacheFetcher) fetch(key string) ([]byte, error) {
	if item := f.get(key); item != nil {
		if time.Since(item.t) <= f.maxAge {
			return item.v, nil
		}
	}

	// As it is, concurrent requests to the cacheFetcher will incur concurrent
	// requests to the fetcher it delegates to.
	val, err := f.fetcher.fetch(key)
	if err != nil {
		return nil, err
	}

	f.set(key, val)
	return val, nil
}

func (f *cacheFetcher) set(key string, val []byte) {
	f.Lock()
	defer f.Unlock()

	f.items[key] = &cacheItem{v: val, t: time.Now()}
}

func (f *cacheFetcher) get(key string) *cacheItem {
	f.RLock()
	defer f.RUnlock()

	if item, ok := f.items[key]; ok {
		return item
	}

	return nil
}
