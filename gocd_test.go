// Copyright (c) 2016 Sebastian Ohm.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"testing"
)

func TestPipelineGroups(t *testing.T) {
	goCD := newGoCD(&fixturesFetcher{})

	if pgs, err := goCD.pipelineGroups(); err == nil {
		if want, have := 2, len(pgs); have != want {
			t.Fatalf("Expected %d got %d", want, have)
		}
	} else {
		t.Fatal(err.Error())
	}
}

func TestPipelineHistory(t *testing.T) {
	goCD := newGoCD(&fixturesFetcher{})

	if phs, err := goCD.pipelineHistory("test-pipeline"); err == nil {
		if want, have := 2, len(phs.Pipelines); have != want {
			t.Fatalf("Expected %d got %d", want, have)
		}
	} else {
		t.Fatal(err.Error())
	}
}

func TestPipelineGroupsError(t *testing.T) {
	tfs := []*fixturesFetcher{
		&fixturesFetcher{fail: true, garbage: false},
		&fixturesFetcher{fail: false, garbage: true},
	}

	for _, tf := range tfs {
		if _, err := newGoCD(tf).pipelineGroups(); err == nil {
			t.FailNow()
		}
	}

	for _, tf := range tfs {
		if _, err := newGoCD(tf).pipelineHistory("test-pipeline"); err == nil {
			t.FailNow()
		}
	}
}
