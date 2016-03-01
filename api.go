// Copyright (c) 2016 Sebastian Ohm.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type apiHandler struct {
	goCD *goCD
}

func newAPIHandler(goCD *goCD) *apiHandler {
	return &apiHandler{
		goCD: goCD,
	}
}

func (a *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {
	case "/api/pipeline_groups.json":
		a.pipelineGroups(w)
	case "/api/pipeline_history.json":
		a.pipelineHistory(w, r.URL.Query())
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (a *apiHandler) pipelineGroups(w http.ResponseWriter) {
	gs, err := a.goCD.pipelineGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	writePipelineGroups(gs, w)
}

func (a *apiHandler) pipelineHistory(w http.ResponseWriter, vals url.Values) {
	names, ok := vals["pipeline"]
	if !ok {
		http.Error(w, "Missing pipeline query parameter", http.StatusBadRequest)
		return
	}

	if len(names) != 1 {
		http.Error(w, "Please specify a single pipeline name", http.StatusBadRequest)
		return
	}

	h, err := a.goCD.pipelineHistory(names[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	writePipelineHistory(h, w)
}

type apiNames []string

func (n apiNames) Len() int           { return len(n) }
func (n apiNames) Less(i, j int) bool { return strings.ToLower(n[i]) < strings.ToLower(n[j]) }
func (n apiNames) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

type apiGroup struct {
	Name      string
	Pipelines apiNames
}

type apiGroups []apiGroup

func (g apiGroups) Len() int           { return len(g) }
func (g apiGroups) Less(i, j int) bool { return strings.ToLower(g[i].Name) < strings.ToLower(g[j].Name) }
func (g apiGroups) Swap(i, j int)      { g[i], g[j] = g[j], g[i] }

func writePipelineGroups(gs []pipelineGroup, w io.Writer) error {
	ags := apiGroups{}

	for _, g := range gs {
		ns := apiNames{}
		for _, p := range g.Pipelines {
			ns = append(ns, p.Name)
		}

		sort.Sort(ns)

		ags = append(ags, apiGroup{Name: g.Name, Pipelines: ns})
	}

	sort.Sort(ags)

	return json.NewEncoder(w).Encode(ags)
}

type apiJob struct {
	Name   string
	Result string
	State  string
}

type apiStage struct {
	Name   string
	Result string
	Jobs   []apiJob
}

type apiPipeline struct {
	Counter int
	Result  string
	Stages  []apiStage
}

func writePipelineHistory(h *pipelineHistory, w io.Writer) error {
	aps := []apiPipeline{}

	for _, p := range h.Pipelines {
		ap := apiPipeline{
			Counter: p.Counter,
			Stages:  []apiStage{},
			Result:  pipelineResult(p.Stages),
		}

		for _, s := range p.Stages {
			js := []apiJob{}

			for _, j := range s.Jobs {
				js = append(js, apiJob{
					Name:   j.Name,
					Result: j.Result,
					State:  j.State,
				})
			}

			ap.Stages = append(ap.Stages, apiStage{
				Name:   s.Name,
				Result: s.Result,
				Jobs:   js,
			})
		}

		aps = append(aps, ap)
	}

	return json.NewEncoder(w).Encode(aps)
}

func pipelineResult(ps []pipelineStage) string {
	for _, p := range ps {
		if p.Result == "Failed" || p.Result == "Unknown" {
			return p.Result
		}
	}

	return "Passed"
}
