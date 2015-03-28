// Copyright 2015 The mRelated Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

/*
Package mRelated extends the multipart package.
*/

package mrelated

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
)

type Reader struct {
	*multipart.Reader
}

// NewReader creates a new mRelated Reader reading from r using the
// given MIME boundary. Actually it is a wrapper for multipart.NewReader
//
// The boundary is usually obtained from the "boundary" parameter of
// the message's "Content-Type" header. Use mime.ParseMediaType to
// parse such headers.
func NewReader(r io.Reader, boundary string) *Reader {
	return &Reader{multipart.NewReader(r, boundary)}
}

// ReadAggregate parses an entire multipart message whose parts are related.
func (r *Reader) ReadAggregate() (o *Aggregate, err error) {
	aggregate := &Aggregate{[]*ObjectHeader{}}

	for {
		p, err := r.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		contentid := p.Header.Get("Content-ID")

		var b bytes.Buffer

		oh := &ObjectHeader{
			ContentId: contentid,
			Header:    p.Header,
		}

		_, err = io.Copy(&b, p)

		oh.content = b.Bytes()

		aggregate.Object = append(aggregate.Object, oh)
	}

	return aggregate, nil
}

// Aggregate is a parsed multipart-related object
type Aggregate struct {
	Object []*ObjectHeader
}

// A ObjectHeader describes a body parts of a multipart-related request
type ObjectHeader struct {
	ContentId string
	Header    textproto.MIMEHeader

	content []byte
}

func (oh *ObjectHeader) Read(p []byte) (n int, err error) {
	r := bytes.NewReader(oh.content)
	n, err = r.Read(p)
	if n >= len(oh.content) {
		return n, io.EOF
	}
	return n, err
}

func (oh *ObjectHeader) String() string {
	return string(oh.content)
}
