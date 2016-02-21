// Copyright (c) 2016 Sebastian Ohm.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssetsHandlerServeHTTP(t *testing.T) {
	type request struct {
		method string
		path   string
		code   int
	}

	rs := []request{
		request{"GET", "/assets/script.js", http.StatusOK},
		request{"GET", "/", http.StatusOK},
		request{"HEAD", "/", http.StatusMethodNotAllowed},
		request{"GET", "/test-pipeline", http.StatusOK},
	}

	for _, r := range rs {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(r.method, r.path, nil)
		if err != nil {
			t.Fatal(err)
		}

		a := newAssetsHandler()
		a.ServeHTTP(rec, req)

		if rec.Code != r.code {
			t.Fatalf("Expected %#v got %d", r, rec.Code)
		}
	}
}
