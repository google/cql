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
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// File storage directory
const uploadDir = ".cql_uploads"

// Mutex for file operations
var fileMutex sync.Mutex

func serve() error {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

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
	
	// File management endpoints
	mux.HandleFunc("/upload_file", handleUploadFile)
	mux.HandleFunc("/delete_file", handleDeleteFile)
	mux.HandleFunc("/list_files", handleListFiles)
	
	// Health check endpoint
	mux.HandleFunc("/health", handleHealthCheck)

	return mux, nil
}

// Type definitions for file management
type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ListFilesResponse struct {
	Files []FileInfo `json:"files"`
}

type DeleteFileRequest struct {
	Filename string `json:"filename"`
}

// handleUploadFile handles file uploads
func handleUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form, 10 MB max
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".cql") {
		http.Error(w, "Only .cql files are allowed", http.StatusBadRequest)
		return
	}

	// Create a new file in the uploads directory
	fileMutex.Lock()
	defer fileMutex.Unlock()

	filePath := filepath.Join(uploadDir, header.Filename)
	
	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get file info for response
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "Failed to get file info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the file size in the response
	response := map[string]int64{
		"size": fileInfo.Size(),
	}
	
	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDeleteFile handles file deletion
func handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Prevent directory traversal
	filename := filepath.Base(req.Filename)
	filePath := filepath.Join(uploadDir, filename)

	fileMutex.Lock()
	defer fileMutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Failed to delete file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleListFiles handles listing all uploaded files
func handleListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		http.Error(w, "Failed to read directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var fileInfos []FileInfo
	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}
		
		// Get file info
		info, err := file.Info()
		if err != nil {
			log.Errorf("Failed to get info for file %s: %v", file.Name(), err)
			continue
		}
		
		fileInfos = append(fileInfos, FileInfo{
			Name: info.Name(),
			Size: info.Size(),
		})
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListFilesResponse{Files: fileInfos})
}

// handleHealthCheck handles health check requests
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
	
	// Read all files from the uploads directory
	uploadedLibraries, err := readUploadedLibraries()
	if err != nil {
		sendError(w, fmt.Errorf("failed to read uploaded libraries: %w", err), http.StatusInternalServerError)
		return
	}
	
	// Add uploaded libraries, FHIRHelpers, and the input CQL to the list of libraries to parse
	libraries := append([]string{evalCQLReq.CQL, fhirHelpers}, uploadedLibraries...)
	
	// Log the libraries being used
	log.Infof("Parsing %d libraries", len(libraries))
	
	elm, err := cql.Parse(req.Context(), libraries, cql.ParseConfig{DataModels: [][]byte{fhirDM}})
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
	w.WriteHeader(code) // Set status code first
	w.Write([]byte("Error: " + err.Error())) // be careful in the future, may not always want to send full error strings to the client
}

type evalCQLRequest struct {
	CQL  string `json:"cql"`
	Data string `json:"data"`
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

// readUploadedLibraries reads all .cql files from the uploads directory
func readUploadedLibraries() ([]string, error) {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	
	// Read all files in the uploads directory
	files, err := os.ReadDir(uploadDir)
	if err != nil {
		if os.IsNotExist(err) {
			// If the directory doesn't exist, return an empty slice
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read uploads directory: %w", err)
	}
	
	var libraries []string
	
	// Read the content of each .cql file
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		// Only process .cql files
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".cql") {
			continue
		}
		
		// Read the file content
		filePath := filepath.Join(uploadDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file.Name(), err)
		}
		
		// Add the file content to the list of libraries
		libraries = append(libraries, string(content))
		log.Infof("Added library from file: %s", file.Name())
	}
	
	return libraries, nil
}
