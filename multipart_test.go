// Copyright 2015 The mRelated Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package mrelated

import (
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
)

var message = `
--example-1
Content-Type: Application/X-FixedRecord
Content-ID: <950120.aaCC@XIson.com>

25
--example-1
Content-Type: Application/octet-stream
Content-Description: The fixed length records
Content-Transfer-Encoding: base64
Content-ID: <950120.aaCB@XIson.com>

T2xkIE1hY0RvbmFsZCBoYWQgYSBmYXJtCkUgSS
BFIEkgTwpBbmQgb24gaGlzIGZhcm0gaGUgaGFk
IHNvbWUgZHVja3MKRSBJIEUgSSBPCldpdGggYS
BxdWFjayBxdWFjayBoZXJlLAphIHF1YWNrIHF1
YWNrIHRoZXJlLApldmVyeSB3aGVyZSBhIHF1YW
NrIHF1YWNrCkUgSSBFIEkgTwo=

--example-1--
`

func TestReadAggregate(t *testing.T) {
	testBody := regexp.MustCompile("\n").ReplaceAllString(message, "\r\n")
	b := strings.NewReader(testBody)
	r := NewReader(b, "example-1")
	a, err := r.ReadAggregate()
	if err != nil {
		t.Fatal("ReadAggregate:", err)
	}

	if g, w := a.Object[0].String(), "25"; g != w {
		t.Errorf("Message = %q, want %q", g, w)
	}

	if g, w := a.Object[0].ContentId, "<950120.aaCC@XIson.com>"; g != w {
		t.Errorf("Message = %q, want %q", g, w)
	}

	if g, w := a.Object[1].String(), "T2xkIE1hY0RvbmFsZCBoYWQgYSBmYXJtCkUgSS\r\n"+
		"BFIEkgTwpBbmQgb24gaGlzIGZhcm0gaGUgaGFk\r\n"+
		"IHNvbWUgZHVja3MKRSBJIEUgSSBPCldpdGggYS\r\n"+
		"BxdWFjayBxdWFjayBoZXJlLAphIHF1YWNrIHF1\r\n"+
		"YWNrIHRoZXJlLApldmVyeSB3aGVyZSBhIHF1YW\r\n"+
		"NrIHF1YWNrCkUgSSBFIEkgTwo=\r\n"; g != w {
		t.Errorf("Message = %q, want %q", g, w)
	}

	if g, w := a.Object[1].ContentId, "<950120.aaCB@XIson.com>"; g != w {
		t.Errorf("Message = %q, want %q", g, w)
	}

	buf, err := ioutil.ReadAll(a.Object[0])
	if err != nil {
		t.Fatal("ReadAll:", err)
	}

	if g, w := string(buf), "25"; g != w {
		t.Errorf("Message = %q, want %q", g, w)
	}
}
