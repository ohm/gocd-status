// Copyright (c) 2016 Sebastian Ohm.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIHandlerServeHTTP(t *testing.T) {
	type request struct {
		method string
		path   string
		code   int
	}

	rs := []request{
		request{"GET", "/api/pipeline_groups.json", http.StatusOK},
		request{"HEAD", "/api/pipeline_groups.json", http.StatusMethodNotAllowed},
		request{"GET", "/api/pipeline_history.json?pipeline=test-pipeline", http.StatusOK},
		request{"GET", "/foo", http.StatusNotFound},
	}

	for _, r := range rs {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(r.method, r.path, nil)
		if err != nil {
			t.Fatal(err)
		}

		a := newAPIHandler(newGoCD(&fixturesFetcher{}))
		a.ServeHTTP(rec, req)

		if rec.Code != r.code {
			t.Fatalf("Expected %#v got %d", r, rec.Code)
		}
	}
}

func TestAPIHandlerResponses(t *testing.T) {
	type response struct {
		path    string
		fixture string
	}

	rs := []response{
		response{"/api/pipeline_groups.json", "api_pipeline_groups.json"},
		response{"/api/pipeline_history.json?pipeline=test-pipeline", "api_pipeline_history.json"},
	}

	for _, r := range rs {
		req, _ := http.NewRequest("GET", r.path, nil)
		rec := httptest.NewRecorder()

		a := newAPIHandler(newGoCD(&fixturesFetcher{}))
		a.ServeHTTP(rec, req)

		body, _ := ioutil.ReadAll(rec.Body)
		if !bytes.Equal(body, loadFixture(r.fixture)) {
			t.FailNow()
		}

		if rec.Header().Get("Content-Type") != "application/json" {
			t.Fail()
		}
	}
}
