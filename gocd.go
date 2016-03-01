// Copyright (c) 2016 Sebastian Ohm.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"encoding/json"
	"fmt"
)

type pipelineJob struct {
	Name   string
	Result string
	State  string
}

type pipelineStage struct {
	Name   string
	Result string
	Jobs   []pipelineJob
}

type pipeline struct {
	Counter int
	Name    string
	Stages  []pipelineStage
}

type pipelineGroup struct {
	Name      string
	Pipelines []pipeline
}

type pipelineHistory struct {
	Pipelines []pipeline
}

type goCD struct {
	fetcher fetcher
}

func newGoCD(fetcher fetcher) *goCD {
	return &goCD{
		fetcher: fetcher,
	}
}

func (g *goCD) pipelineGroups() ([]pipelineGroup, error) {
	val, err := g.fetcher.fetch("/go/api/config/pipeline_groups")
	if err != nil {
		return nil, err
	}

	pgs := []pipelineGroup{}
	if err := json.Unmarshal(val, &pgs); err != nil {
		return nil, err
	}

	return pgs, nil
}

func (g *goCD) pipelineHistory(name string) (*pipelineHistory, error) {
	val, err := g.fetcher.fetch(fmt.Sprintf("/go/api/pipelines/%s/history", name))
	if err != nil {
		return nil, err
	}

	phs := &pipelineHistory{}
	if err := json.Unmarshal(val, &phs); err != nil {
		return nil, err
	}

	return phs, nil
}
