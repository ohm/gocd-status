// Copyright (c) 2016 Sebastian Ohm.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestHttpFetcher(t *testing.T) {
	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/good":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "OK")
			case "/bad":
				http.Error(w, "404", http.StatusNotFound)
			default:
				t.Fatalf(fmt.Sprintf("%v not implemented", r.URL))
			}
		}),
	)

	f, _ := newHTTPFetcher(s.URL)

	if r, err := f.fetch("/good"); err == nil {
		if string(r) != "OK\n" {
			t.Fatalf("Expected fetching /good body to be \"OK\": %v", r)
		}
	} else {
		t.Fatalf("Expected to be able to fetch /good")
	}

	if _, err := f.fetch("/bad"); err == nil {
		t.Fatalf("Expected fetching /bad to fail")
	}
}

type fakeFetcher struct {
	counts map[string]int
}

func newFakeFetcher() *fakeFetcher {
	return &fakeFetcher{
		counts: map[string]int{},
	}
}

func (f *fakeFetcher) fetch(key string) ([]byte, error) {
	if key == "fail" {
		return nil, fmt.Errorf("Unable to fetch key %s", key)
	}

	if v, ok := f.counts[key]; ok {
		f.counts[key] = v + 1
		return []byte(strconv.Itoa(v + 1)), nil
	}

	f.counts[key] = 1
	return []byte("1"), nil
}

func TestFakeFetcher(t *testing.T) {
	f := newFakeFetcher()

	if _, err := f.fetch("fail"); err == nil {
		t.Fatalf("Expected fetching fail to fail")
	}

	for i := 1; i < 10; i++ {
		if v, err := f.fetch("key1"); err == nil {
			if j, _ := strconv.Atoi(string(v)); j != i {
				t.Fatalf("Expected %d to be %d", j, i)
			}
		} else {
			t.Fatalf("Unable to fetch key1: %s", err.Error())
		}
	}
}

func TestCacheFetcherCaches(t *testing.T) {
	f, _ := newCacheFetcher(newFakeFetcher(), 1*time.Second, 1*time.Second, 1)

	for i := 1; i < 10; i++ {
		v, _ := f.fetch("key2")
		if j, _ := strconv.Atoi(string(v)); j != 1 {
			t.Fatalf("Expected fetching %d to be %d", j, i)
		}
	}
}

func TestCacheFetcherExpires(t *testing.T) {
	f, _ := newCacheFetcher(newFakeFetcher(), 0*time.Second, 1*time.Second, 2)

	for i := 1; i < 10; i++ {
		v, _ := f.fetch("key3")
		if j, _ := strconv.Atoi(string(v)); j != i {
			t.Fatalf("Expected fetching %d to be %d", j, i)
		}
	}
}

func TestCacheFetcherPropagatesError(t *testing.T) {
	f, _ := newCacheFetcher(newFakeFetcher(), 1*time.Second, 1*time.Second, 3)

	if _, err := f.fetch("fail"); err == nil {
		t.Fatalf("Expected fetching fail to fail")
	}
}

const fixturesDir = "./_fixtures"

type fixturesFetcher struct {
	fail    bool
	garbage bool
}

func (t *fixturesFetcher) fetch(key string) ([]byte, error) {
	if t.fail == true {
		return nil, errors.New("Fail")
	}

	if t.garbage == true {
		return []byte("Not JSON"), nil
	}

	switch key {
	case "/go/api/config/pipeline_groups":
		return loadFixture("pipeline_groups.json"), nil
	case "/go/api/pipelines/test-pipeline/history":
		return loadFixture("pipeline_history.json"), nil
	}

	panic("Not implemented")
}

func loadFixture(f string) []byte {
	if r, err := ioutil.ReadFile(filepath.Join(fixturesDir, f)); err == nil {
		return r
	}

	panic("Unable to load fixture:" + f)
}
