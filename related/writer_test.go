// Copyright 2015 The multipart-related Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package related

import (
	"bytes"
	"mime"
	"net/textproto"
	"reflect"
	"testing"
)

func TestWriter(t *testing.T) {
	fileContents := []byte(`Life? Don't talk to me about life!`)

	var b bytes.Buffer
	w := NewWriter(&b)
	{
		part, err := w.CreateRoot("", "", nil)
		if err != nil {
			t.Fatalf("CreateRoot: %v", err)
		}
		part.Write(fileContents)

		if err := w.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}

		s := b.String()
		if len(s) == 0 {
			t.Fatal("String: unexpected empty result")
		}
	}

	// NewReader(&b, w.Boundary())
	// TODO Check output
}

func TestCreateRootFail(t *testing.T) {
	var b bytes.Buffer
	w := NewWriter(&b)

	// Error handling
	testsError := []struct {
		id    string
		media string
	}{
		{"dont", "dont/panic"},
		{"", "dont;panic"},
		{"dont", ""},
	}

	for i, tt := range testsError {
		if _, err := w.CreateRoot(tt.id, tt.media, nil); err == nil {
			t.Errorf("%d. Content-Id: %s, Media-Type: %s", i, tt.id, tt.media)
		}
	}

	for i := 2; i > 0; i-- {
		_, err := w.CreateRoot("", "a/b", nil)
		if i == 1 && err != ErrRootExists {
			t.Errorf("%d. Multiple CreateRoot: Expected error", i)
		}
	}

	w.Close()
}

func TestCreatePartFail(t *testing.T) {
	var b bytes.Buffer
	w := NewWriter(&b)

	// Error handling
	id := "&&;&;&"
	if _, err := w.CreatePart(id, nil); err == nil {
		t.Errorf("Content-Id: %s", id)
	}

	w.Close()
}

func TestCreatePartFirst(t *testing.T) {
	h := textproto.MIMEHeader{}
	h.Add("Content-Type", "a/b")

	tests := []struct {
		id        string
		header    textproto.MIMEHeader
		mediaType string
	}{
		{"a@b.c", nil, DefaultMediaType},
		{"a@b.c", h, "a/b"},
	}

	for i, tt := range tests {
		var b bytes.Buffer
		w := NewWriter(&b)

		if w.firstPart != false {
			t.Errorf("Before:\n%d. firstPart = %t, want %t", i, w.firstPart, false)
		}

		if _, err := w.CreatePart(tt.id, tt.header); err != nil {
			t.Fatalf("%d. CreatePart: %v", i, err)
		}
		if w.mediaType != tt.mediaType {
			t.Errorf("%d. type = %s, want %s", i, w.mediaType, tt.mediaType)
		}
		if w.rootMediaType != tt.mediaType {
			t.Errorf("%d. type = %s, want %s", i, w.rootMediaType, tt.mediaType)
		}

		if w.firstPart != true {
			t.Errorf("After:\n%d. firstPart = %t, want %t", i, w.firstPart, true)
		}

		w.Close()
	}
}

func TestSetStart(t *testing.T) {
	var b bytes.Buffer
	w := NewWriter(&b)
	tests := []struct {
		id string
		w  string
		ok bool
	}{
		{"a@b.c", "<a@b.c>", true},
		{"aa", "", false},
	}

	for i, tt := range tests {
		err := w.SetStart(tt.id)
		got := err == nil
		if got != tt.ok {
			t.Errorf("%d. start %q = %v (%v), want %v", i, tt.id, got, err, tt.ok)
		} else if tt.ok {
			got := w.start
			if got != tt.w {
				t.Errorf("start = %q; want %q", got, tt.w)
			}
		}
	}
	w.Close()
}

func TestSetType(t *testing.T) {
	var b bytes.Buffer
	w := NewWriter(&b)
	tests := []struct {
		t  string
		ok bool
	}{
		{"application/json", true},
		{";", false},
	}

	for i, tt := range tests {
		err := w.SetType(tt.t)
		got := err == nil
		if got != tt.ok {
			t.Errorf("%d. start %q = %v (%v), want %v", i, tt.t, got, err, tt.ok)
		} else if tt.ok {
			got := w.mediaType
			if got != tt.t {
				t.Errorf("start = %q; want %q", got, tt.t)
			}
		}
	}
	w.Close()
}

func TestSetBoundary(t *testing.T) {
	var b bytes.Buffer
	w := NewWriter(&b)
	tests := []struct {
		b  string
		ok bool
	}{
		{"abc", true},
		{"ungültig", false},
	}

	for i, tt := range tests {
		err := w.SetBoundary(tt.b)
		got := err == nil
		if got != tt.ok {
			t.Errorf("%d. start %q = %v (%v), want %v", i, tt.b, got, err, tt.ok)
		} else if tt.ok {
			got := w.Boundary()
			if got != tt.b {
				t.Errorf("start = %q; want %q", got, tt.b)
			}
		}
	}

	w.Close()
}

func TestClose(t *testing.T) {
	var b bytes.Buffer
	w := NewWriter(&b)

	if _, err := w.CreateRoot("a@b.c", "text/plain", nil); err != nil {
		t.Fatalf("CreateRoot: %v", err)
	}
	if err := w.SetType("text/html"); err != nil {
		t.Fatalf("SetType: %v", err)
	}
	if err := w.Close(); err != ErrTypeMatch {
		t.Errorf("NoMediaType = %v; want %q", err, ErrTypeMatch)
	}
	w.Close()
}

func TestFormDataContentType(t *testing.T) {
	var b bytes.Buffer

	in := map[string]string{
		"boundary":   "abc",
		"type":       "text/plain",
		"start":      "a@b.c",
		"start-info": `-o p"s`,
	}
	want := map[string]string{
		"boundary":   "abc",
		"type":       "text/plain",
		"start":      "<a@b.c>",
		"start-info": `-o p\"s`,
	}

	w := NewWriter(&b)

	if err := w.SetBoundary(in["boundary"]); err != nil {
		t.Fatalf("SetBoundary: %v", err)
	}
	if err := w.SetType(in["type"]); err != nil {
		t.Fatalf("SetType: %v", err)
	}
	if err := w.SetStart(in["start"]); err != nil {
		t.Fatalf("SetStart: %v", err)
	}
	w.SetStartInfo(in["start-info"])

	got := w.FormDataContentType()
	mediatype, params, err := mime.ParseMediaType(got)
	if err != nil {
		t.Fatalf("ParseMediaType: %v", err)
	}
	if mediatype != "multipart/related" {
		t.Errorf("mediatype = %s, want multipart/related", mediatype)
	}
	if !reflect.DeepEqual(params, want) {
		t.Errorf("params = %v, want %v", params, want)
	}

	w.Close()
}

func TestFormatContentId(t *testing.T) {
	tests := []struct {
		id string
		w  string
		ok bool
	}{
		{"a@b.c", "<a@b.c>", true},
		{"<aa>", "", false},
		{"", "", false},
	}

	for i, tt := range tests {
		got, err := formatContentId(tt.id)
		if err == nil != tt.ok {
			t.Errorf("%d. start %q = %v (%v), want %v", i, tt.id, got, err, tt.ok)
		} else if tt.ok {
			if got != tt.w {
				t.Errorf("start = %q; want %q", got, tt.w)
			}
		}
	}
}
