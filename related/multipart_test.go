// Copyright 2015 The multipart-related Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package related

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

var testParams = map[string]string{
	"boundary": "example-1",
	"start":    "a@b.c",
	"type":     "a/b",
}

var testBody = `--example-1
Content-Type: a/b
Content-ID: <a@b.c>

Life?
--example-1
Content-Type: b/c
Content-Transfer-Encoding: base64
Content-ID: <b@c.d>

RG9uJ3QgdGFsayB0byBtZSBhYm91dCBsaWZlIQ==

--example-1--`

func TestReader(t *testing.T) {
	r := strings.NewReader(testBody)
	reader := NewReader(r, testParams)
	buf := new(bytes.Buffer)

	// Part 1
	part, err := reader.NextPart()
	if part == nil || err != nil {
		t.Error("Expected part1")
		return
	}
	buf.Reset()
	if _, err := io.Copy(buf, part); err != nil {
		t.Errorf("part 1 copy: %v", err)
	}

	if g, w := buf.String(), "Life?"; g != w {
		t.Errorf("part 1 (%q) body = %s, want %s", "\r\nLife?\n", g, w)
	}

	// Part 2
	part, err = reader.NextPart()
	if err != nil {
		t.Error("Expected part2")
		return
	}
	buf.Reset()
	if _, err := io.Copy(buf, part); err != nil {
		t.Errorf("part 2 copy: %v", err)
	}
	gotBody := "\r\nRG9uJ3QgdGFsayB0byBtZSBhYm91dCBsaWZlIQ==\n"
	if g, w := buf.String(), "Don't talk to me about life!"; g != w {
		t.Errorf("part 2 (%q) body = %s, want %s", gotBody, g, w)
	}

	// Non-existent Part 3
	part, err = reader.NextPart()
	if part != nil {
		t.Error("Didn't expected part 3.")
		return
	}
	if err != io.EOF {
		t.Errorf("part 3 expected io.EOF; got %v", err)
	}
}

var testDupBody = `--example-1
Content-Type: a/b
Content-ID: <a@b.c>

Life?
--example-1
Content-Type: b/c
Content-Transfer-Encoding: base64
Content-ID: <a@b.c>

RG9uJ3QgdGFsayB0byBtZSBhYm91dCBsaWZlIQ==

--example-1--`

func TestDuplicateRoots(t *testing.T) {
	r := strings.NewReader(testDupBody)
	reader := NewReader(r, testParams)

	// Part 1
	part, err := reader.NextPart()
	if part == nil || err != nil {
		t.Error("Expected part1")
		return
	}
	// Part 2
	part, err = reader.NextPart()
	if err != ErrDupRoot {
		t.Errorf("Expected error = %v, want %v", err, ErrDupRoot)
	}
}

var testMovedRootBody = `--example-1
Content-Type: a/b
Content-ID: <b@c.d>

Life?
--example-1
Content-Type: b/c
Content-Transfer-Encoding: base64
Content-ID: <a@b.c>

RG9uJ3QgdGFsayB0byBtZSBhYm91dCBsaWZlIQ==

--example-1--`

func TestMovedRoot(t *testing.T) {
	r := strings.NewReader(testMovedRootBody)
	reader := NewReader(r, testParams)

	// Part 1
	part, err := reader.NextPart()
	if part == nil || err != nil {
		t.Error("Expected part1")
		return
	}
	if part.Root == true {
		t.Errorf("Part 1 root = %t, want %t", part.Root, false)
	}
	// Part 2
	part, err = reader.NextPart()
	if part.Root == false {
		t.Errorf("Part 2 root = %t, want %t", part.Root, true)
	}
}

var testParamsWithOutStart = map[string]string{
	"boundary": "example-1",
	"type":     "a/b",
}
var testFirstPartBody = `--example-1
Content-Type: a/b
Content-ID: <b@c.d>

Life?
--example-1
Content-Type: b/c
Content-Transfer-Encoding: base64
Content-ID: <a@b.c>

RG9uJ3QgdGFsayB0byBtZSBhYm91dCBsaWZlIQ==

--example-1--`

func TestFirstPartRoot(t *testing.T) {
	r := strings.NewReader(testMovedRootBody)
	reader := NewReader(r, testParamsWithOutStart)

	// Part 1
	part, err := reader.NextPart()
	if part == nil || err != nil {
		t.Error("Expected part1")
		return
	}
	if part.Root == false {
		t.Errorf("Part 1 root = %t, want %t", part.Root, false)
	}
	// Part 2
	part, err = reader.NextPart()
	if part.Root == true {
		t.Errorf("Part 2 root = %t, want %t", part.Root, false)
	}
}

func TestReadObject(t *testing.T) {
	r := strings.NewReader(testMovedRootBody)
	reader := NewReader(r, testParamsWithOutStart)

	// Part 1
	object, err := reader.ReadObject()
	if object.Values == nil || err != nil {
		t.Error("Object", err)
		return
	}

	if want := `Life?`; string(object.Values[0].content) != want {
		t.Errorf("Object 1 body = %q, want %q", object.Values[0].content, want)
	}
	want := `Don't talk to me about life!`
	if string(object.Values[1].content) != want {
		t.Errorf("Object 2 body = %q, want %q", object.Values[1].content, want)
	}
}

func TestMovedReadObject(t *testing.T) {
	r := strings.NewReader(testMovedRootBody)
	reader := NewReader(r, testParams)

	// Part 1
	object, err := reader.ReadObject()
	if object.Values == nil || err != nil {
		t.Error("Object", err)
		return
	}

	want := `Don't talk to me about life!`
	if string(object.Values[0].content) != want {
		t.Errorf("Object 1 body = %q, want %q", object.Values[0].content, want)
	}

	if want := `Life?`; string(object.Values[1].content) != want {
		t.Errorf("Object 2 body = %q, want %q", object.Values[1].content, want)
	}
}

func TestParseContentId(t *testing.T) {
	tests := []struct {
		id string
		w  string
	}{
		{"<a@b.c>", "a@b.c"},
		{"<aa>", ""},
		{"", ""},
	}

	for i, tt := range tests {
		got := parseContentId(tt.id)
		if got != tt.w {
			t.Errorf("%d. parseContent(%s) = %s; want %s", i, tt.id, got, tt.w)
		}
	}
}
