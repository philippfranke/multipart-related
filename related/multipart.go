// Copyright 2015 The multipart-related Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package related

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"mime/multipart"
	"net/mail"
	"net/textproto"
)

var (
	ErrDupRoot = errors.New("Detect duplicate roots")
)

// Reader is an iterator over parts in a MIME multipart/related body.
type Reader struct {
	// SkipMatch controls whether a Reader matches the root body part's
	// content-type against compound object's type
	// SkipMatch bool

	// start is the content-ID of the compound object's "root"; optional
	start string

	// mediaType is the MIME media type of the compound object; required
	mediaType string

	// startInfo provides additional information to an application
	startInfo string

	r        *multipart.Reader
	rootRead bool
}

// NewReader returns a new multipart/related Reader reading from r using the
// given MIME boundary. It's a wrapper around multipart's Reader
func NewReader(
	r io.Reader,
	params map[string]string,
) *Reader {
	return &Reader{
		r:         multipart.NewReader(r, params["boundary"]),
		mediaType: params["type"],
		start:     parseContentId(params["start"]),
		startInfo: params["start-info"],
		rootRead:  false,
	}
}

// A Part represents a single part in a multipart/related body
type Part struct {
	Header textproto.MIMEHeader
	Root   bool

	// r is either a reader directly reading from p, or it's a wrapper
	// around such a reader, decoding the Content-Tranfer-Encoding
	r io.Reader
}

// A Object is parsed multipart/related compound object.
type Object struct {
	Values []*ObjectHeader
}

// A ObjectHeader describes a component of the aggregate whole of a
// multipart/related request.
type ObjectHeader struct {
	content []byte
	i       int64 // current reading index
}

// NextPart returns the next part in the multipart/related or and error.
// When there are no more parts, the error io.EOF is returned.
func (r *Reader) NextPart() (*Part, error) {
	wrap, err := r.r.NextPart()
	if err != nil {
		return nil, err
	}
	p := &Part{
		Header: wrap.Header,
		Root:   false,
	}

	contentId := parseContentId(p.Header.Get("Content-Id"))
	if r.start != "" && r.start == contentId {
		if r.rootRead {
			return nil, ErrDupRoot
		} else {
			p.Root = true
			r.rootRead = true
		}
	} else if !r.rootRead && r.start == "" {
		p.Root = true
		r.rootRead = true
	}
	p.r = wrap

	switch p.Header.Get("Content-Transfer-Encoding") {
	case "base64":
		p.Header.Del("Content-Transfer-Encoding")
		p.r = base64.NewDecoder(base64.StdEncoding, p.r)
		break
		// TODO binary reader
	}
	// TODO SkipMatch

	return p, nil
}

// Read reads the body of a part, after its headers and before the next
// part (if any) begins. It's a wrapper around multipart's Part.Read()
func (p *Part) Read(d []byte) (n int, err error) {
	return p.r.Read(d)
}

// ReadObject parses an entire multipart/related message.
func (r *Reader) ReadObject() (*Object, error) {
	object := &Object{[]*ObjectHeader{}}
	for {
		p, err := r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		var b bytes.Buffer

		if _, err := io.Copy(&b, p); err != nil {
			return nil, err
		}

		oh := &ObjectHeader{
			content: b.Bytes(),
		}
		if p.Root {
			object.Values = append([]*ObjectHeader{oh}, object.Values...)
		} else {
			object.Values = append(object.Values, oh)
		}

	}

	return object, nil
}

// Read reads the content of a ObjectHeader.
func (oh *ObjectHeader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if oh.i >= int64(len(oh.content)) {
		return 0, io.EOF
	}
	n = copy(b, oh.content[oh.i:])
	oh.i += int64(n)
	return
}

func parseContentId(contentId string) string {
	addr, err := mail.ParseAddress(contentId)
	if err != nil {
		return ""
	}
	return addr.Address
}
