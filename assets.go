//go:generate sh -c "m4 assets_gen.go.m4 | gofmt > assets_gen.go"
// Copyright (c) 2016 Sebastian Ohm.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"net/http"
)

type assetsHandler struct{}

func newAssetsHandler() *assetsHandler {
	return &assetsHandler{}
}

func (h *assetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Not allowed", http.StatusMethodNotAllowed)
	}

	w.WriteHeader(http.StatusOK)

	switch r.URL.Path {
	case "/assets/script.js":
		w.Header().Add("Content-Type", "application/javascript")
		w.Write(assetScriptJS)
	default:
		w.Header().Add("Content-Type", "text/html")
		w.Write(assetIndexHTML)
	}
}
