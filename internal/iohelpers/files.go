// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package iohelpers contains functions for file I/O both locally and in GCS.
package iohelpers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"github.com/google/bulk_fhir_tools/gcs"
)

// IOConfig contains configuration options for IO functions.
type IOConfig struct {
	GCSEndpoint string
}

// FilesWithSuffix returns all file paths in a directory that end with a given suffix.
func FilesWithSuffix(ctx context.Context, dir string, suffix string, cfg *IOConfig) ([]string, error) {
	if strings.HasPrefix(dir, "gs://") {
		if cfg == nil {
			return nil, fmt.Errorf("FilesWithSuffix() IOConfig cannot be nil for GCS paths, but was nil. path: %s", dir)
		}
		return filesWithSuffixGCS(ctx, dir, suffix, *cfg)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	filePaths := []string{}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), suffix) {
			continue
		}
		filePaths = append(filePaths, filepath.Join(dir, file.Name()))
	}
	return filePaths, nil
}

// filesWithSuffixGCS returns all file paths in a GCS bucket that end with a given suffix.
func filesWithSuffixGCS(ctx context.Context, gcsPath, suffix string, cfg IOConfig) ([]string, error) {
	filePaths := []string{}
	bucket, path, err := gcs.PathComponents(gcsPath)
	if err != nil {
		return nil, err
	}
	client, err := gcs.NewClient(ctx, bucket, cfg.GCSEndpoint)
	if err != nil {
		return nil, fmt.Errorf("could not connect to a GCS client %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("error closing GCS client: %v", err)
		}
	}()

	bucketHandle := client.Bucket(bucket)
	iter := bucketHandle.Objects(ctx, &storage.Query{Prefix: path, Delimiter: "/"})
	for {
		obj, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(obj.Name, suffix) {
			filePaths = append(filePaths, "gs://"+bucket+"/"+obj.Name)
		}
	}
	return filePaths, nil
}

// ReadFile reads the contents of a file at the given path.
// If the path is a GCS it will attempt to read the file from GCS.
func ReadFile(ctx context.Context, filePath string, cfg *IOConfig) (contents []byte, funcErr error) {
	if strings.HasPrefix(filePath, "gs://") {
		if cfg == nil {
			return nil, fmt.Errorf("ReadFile() IOConfig cannot be nil for GCS paths, but was nil. path: %s", filePath)
		}
		return readGCSFile(ctx, filePath, *cfg)
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = f.Close(); err != nil {
			if funcErr == nil {
				funcErr = err
			}
		}
	}()

	return io.ReadAll(f)
}

// readGCSFile reads the contents of a file at the given GCS path.
func readGCSFile(ctx context.Context, gcsPath string, cfg IOConfig) ([]byte, error) {
	bucket, objPath, err := gcs.PathComponents(gcsPath)
	if err != nil {
		return nil, fmt.Errorf("could not parse GCS path %q: %w", gcsPath, err)
	}
	client, err := gcs.NewClient(ctx, bucket, cfg.GCSEndpoint)
	if err != nil {
		return nil, fmt.Errorf("could not connect to a GCS client %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("error closing GCS client: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	r, err := client.GetFileReader(ctx, objPath)
	if err != nil {
		return nil, fmt.Errorf("could not access reader for %s/%s: %w", bucket, objPath, err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			fmt.Printf("error closing GCS file reader: %v", err)
		}
	}()

	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll failed to read GCS file %s/%s: %w", bucket, objPath, err)
	}
	return content, nil
}

// WriteFile writes the given content to a file at the given path.
// If the path is a GCS it will attempt to write the file to GCS.
func WriteFile(ctx context.Context, filePath, fileName string, content []byte, cfg *IOConfig) error {
	if strings.HasPrefix(filePath, "gs://") {
		if cfg == nil {
			return fmt.Errorf("WriteFile() IOConfig cannot be nil for GCS paths, but was nil. path: %s", filePath)
		}
		return writeGCSFile(ctx, filePath, fileName, content, *cfg)
	}
	return os.WriteFile(path.Join(filePath, fileName), content, 0644)
}

func writeGCSFile(ctx context.Context, gcsPath, fileName string, content []byte, cfg IOConfig) error {
	fullFilePath := gcsPath + "/" + fileName
	bucket, objPath, err := gcs.PathComponents(fullFilePath)
	if err != nil {
		return fmt.Errorf("could not parse GCS path %q: %w", fullFilePath, err)
	}
	client, err := gcs.NewClient(ctx, bucket, cfg.GCSEndpoint)
	if err != nil {
		return fmt.Errorf("could not connect to a GCS client %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("error closing GCS client: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	w := client.GetFileWriter(ctx, objPath)
	defer func() {
		if err := w.Close(); err != nil {
			fmt.Printf("error closing GCS file writer: %v", err)
		}
	}()

	_, err = w.Write(content)
	return err
}
