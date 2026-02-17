// Copyright 2020-2025 Buf Technologies, Inc.
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

// Package tmp provides temporary file and directory utilities.
package tmp

import (
	"context"
	"io"
	"os"
)

// Dir is a temporary directory.
type Dir interface {
	// Path returns the path of the temporary directory.
	Path() string
	// Close removes the temporary directory and all its contents.
	Close() error
}

// File is a temporary file.
type File interface {
	// Path returns the path of the temporary file.
	Path() string
	// Close removes the temporary file.
	Close() error
}

type dir struct {
	path string
}

type file struct {
	path string
}

// NewDir creates a new temporary directory.
//
// The caller must call Close to remove the directory when done.
func NewDir(_ context.Context) (Dir, error) {
	path, err := os.MkdirTemp("", "buf-tmp-*")
	if err != nil {
		return nil, err
	}
	return &dir{path: path}, nil
}

// NewFile creates a new temporary file with the contents of r.
//
// The caller must call Close to remove the file when done.
func NewFile(_ context.Context, r io.Reader) (File, error) {
	f, err := os.CreateTemp("", "buf-tmp-*")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return nil, err
	}
	return &file{path: f.Name()}, nil
}

func (d *dir) Path() string {
	return d.path
}

func (d *dir) Close() error {
	return os.RemoveAll(d.path)
}

func (f *file) Path() string {
	return f.path
}

func (f *file) Close() error {
	return os.Remove(f.path)
}
