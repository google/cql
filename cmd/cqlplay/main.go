// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// cqlplay is an experimental	Go program that serves our CQL playground on localhost:8080. This is
// NOT meant to be run in production, but is a useful tool for playing with and testing our CQL
// engine.
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"time"

	"flag"

	log "github.com/golang/glog"
	"github.com/google/cql"
	"github.com/google/cql/retriever/local"
	"github.com/google/cql/terminology"
)

//go:embed static/*
var staticAssets embed.FS

//go:embed testdata/terminology/*.json
var terminologyDir embed.FS

func main() {
	flag.Parse()
	if err := serve(); err != nil {
		log.Fatalf("cqlplay failed with an error: %v", err)
	}

}

// tp is a shared terminology provider. This must be thread safe.
var tp *terminology.LocalFHIRProvider

func serve() error {
	mux, err := serverHandler()
	if err != nil {
		return err
	}
	fmt.Println("serving on port 8080, try http://localhost:8080")
	return http.ListenAndServe("localhost:8080", mux)
}

func serverHandler() (http.Handler, error) {
	var err error
	tp, err = getTerminologyProvider()
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	// Serve the static assets from the static/ folder as the root of the fileserver.
	staticFS, err := fs.Sub(staticAssets, "static")
	if err != nil {
		return nil, err
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// eval_cql is the evaluation endpoint for CQL.
	mux.HandleFunc("/eval_cql", handleEvalCQL)

	return mux, nil
}

func handleEvalCQL(w http.ResponseWriter, req *http.Request) {
	// 5MB limit for body size:
	bodyR := io.LimitReader(req.Body, 5e6)
	inputCQL, err := io.ReadAll(bodyR)
	if err != nil {
		sendError(w, err, http.StatusInternalServerError)
		return
	}
	evalCQLReq := &evalCQLRequest{}
	if err := json.Unmarshal(inputCQL, evalCQLReq); err != nil {
		sendError(w, err, http.StatusInternalServerError)
		return
	}
	log.Infof("Request: %+v", evalCQLReq)

	fhirDM, err := cql.FHIRDataModel("4.0.1")
	if err != nil {
		sendError(w, err, http.StatusInternalServerError)
		return
	}
	fhirHelpers, err := cql.FHIRHelpersLib("4.0.1")
	if err != nil {
		sendError(w, err, http.StatusInternalServerError)
		return
	}

	// Combine main CQL with additional libraries and FHIRHelpers
	cqlInputs := append([]string{evalCQLReq.CQL}, evalCQLReq.Libraries...)
	// TODO: now that users can supply their own libraries, we may wish to only add FHIRHelpers if
	// it's not already included. Though, our parser really only works with FHIR Helpers 4.0.1.
	cqlInputs = append(cqlInputs, fhirHelpers)

	elm, err := cql.Parse(req.Context(), cqlInputs, cql.ParseConfig{DataModels: [][]byte{fhirDM}})
	if err != nil {
		sendError(w, fmt.Errorf("failed to parse: %w", err), http.StatusInternalServerError)
		return
	}

	var ret *local.Retriever
	if evalCQLReq.Data != "" {
		ret, err = local.NewRetrieverFromR4Bundle([]byte(evalCQLReq.Data))
		if err != nil {
			sendError(w, fmt.Errorf("unable to load patient bundle: %w", err), http.StatusInternalServerError)
			return
		}
	}

	start := time.Now()
	results, err := elm.Eval(req.Context(), ret, cql.EvalConfig{Terminology: tp})
	if err != nil {
		sendError(w, fmt.Errorf("failed to eval: %w", err), http.StatusInternalServerError)
		return
	}

	evalTime := time.Since(start)
	log.Infof("eval time: %v", evalTime)

	resJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		sendError(w, fmt.Errorf("unable to marshal CQL response: %w", err), http.StatusInternalServerError)
		return
	}
	w.Write(resJSON)
}

func sendError(w http.ResponseWriter, err error, code int) {
	log.Errorf("%v", err)
	w.Write([]byte("Error: " + err.Error())) // be careful in the future, may not always want to send full error strings to the client
	w.WriteHeader(code)
}

type evalCQLRequest struct {
	CQL       string   `json:"cql"`
	Data      string   `json:"data"`
	Libraries []string `json:"libraries"`
}

func getTerminologyProvider() (*terminology.LocalFHIRProvider, error) {
	entries, err := terminologyDir.ReadDir("testdata/terminology")
	if err != nil {
		return nil, err
	}

	var valuesets = make([]string, 0, len(entries))
	for _, entry := range entries {
		eData, err := terminologyDir.ReadFile("testdata/terminology/" + entry.Name())
		if err != nil {
			return nil, err
		}
		valuesets = append(valuesets, string(eData))
	}
	tp, err := terminology.NewInMemoryFHIRProvider(valuesets)
	if err != nil {
		return nil, err
	}
	return tp, nil
}
