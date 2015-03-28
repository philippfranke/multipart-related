// Copyright 2015 The mRelated Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package mRelated

import (
	"io"
	"mime/multipart"
)

type Writer struct {
	*multipart.Writer
}

// NewWriter returns a new multipart Writer with a random boundary,
// writing to w. Actually it is a wrapper for multipart.NewWriter
func NewWriter(w io.Writer) *Writer {
	return &Writer{multipart.NewWriter(w)}
}

// FormDataContentType returns the Content-Type for a
// multipart/related with this Writer's Boundary.
func (w *Writer) FormDataContentType() string {
	return "multipart/related; boundary=" + w.Boundary()
}
