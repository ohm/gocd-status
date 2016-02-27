// Copyright (c) 2016 Sebastian Ohm.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	var (
		listen  = flag.String("listen", ":8080", "HTTP listen address")
		maxAge  = flag.Duration("max-age", 15*time.Second, "How long to cache GoCD API responses")
		maxWait = flag.Duration("max-wait", 1*time.Second, "How long to wait for GoCD API responses")
		maxReqs = flag.Int("max-requests", 1, "How many concurrent requests to issue against GoCD API")
		url     = flag.String("url", "https://user:pass@ci.example.com/", "Address of the GoCD API")
	)
	flag.Parse()

	hf, err := newHTTPFetcher(*url)
	if err != nil {
		log.Fatal(err)
	}

	cf, err := newCacheFetcher(hf, *maxAge, *maxWait, *maxReqs)
	if err != nil {
		log.Fatal(err)
	}

	goCD := newGoCD(cf)

	s := http.NewServeMux()
	s.Handle("/api/", newAPIHandler(goCD))
	s.Handle("/", newAssetsHandler())

	log.Fatal(http.ListenAndServe(*listen, s))
}
