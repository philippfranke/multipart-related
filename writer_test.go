// Copyright 2015 The mRelated Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package mrelated

import (
	"bytes"
	"testing"
)

func TestFormDataContentType(t *testing.T) {
	var b bytes.Buffer
	w := NewWriter(&b)
	part, err := w.CreatePart(nil)
	if err != nil {
		t.Fatal("CreatePart:", err)
	}
	part.Write([]byte("Test"))

	if err := w.Close(); err != nil {
		t.Fatal("Close:", err)
	}
	w.SetBoundary("test")
	g := w.FormDataContentType()
	e := "multipart/related; boundary=test"
	if g != e {
		t.Errorf("Content-Type = %q, want %q", g, e)
	}
}
