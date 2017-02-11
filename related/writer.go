// Copyright 2015 The multipart-related Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package related

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	//"net/mail"
	"net/textproto"
	"strings"
)

// DefaultMediaType
const DefaultMediaType = "text/plain; charset=utf-8"

// Errors introduced by the multipart/related.
var (
	ErrTypeMatch  = errors.New("root's media type doesn't match")
	ErrRootExists = errors.New("root part already exists")
)

// A Writer generates multipart/related messages.
// See http://tools.ietf.org/html/rfc2387
type Writer struct {
	w *multipart.Writer

	// start is the content-ID of the compound object's "root"; optional
	start string

	// mediaType is the MIME media type of the compound object; required
	mediaType string

	// startInfo provides additional information to an application
	startInfo string

	// rootMediaType is the MIME media type of the "root" part
	rootMediaType string

	// Used for setting compound object's media-type without CreateRoot
	firstPart bool

	// Prevent multiple CreateRoot calls
	rootPart bool
}

// NewWriter returns a new multipart/related Writer with a random
// boundary, writing to w. It's a wrapper around multipart's Writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:         multipart.NewWriter(w),
		firstPart: false,
		rootPart:  false,
	}
}

// Boundary is a wrapper around multipart's Writer.Boundary
func (w *Writer) Boundary() string {
	return w.w.Boundary()
}

// SetBoundary is a wrapper around multipart's Writer.SetBoundary
//
// SetBoundary overrides the Writer's default randomly-generated
// boundary separator with an explicit value.
//
// SetBoundary must be called before any parts are created, may only
// contain certain ASCII characters, and must be 1-69 bytes long.
func (w *Writer) SetBoundary(boundary string) error {
	return w.w.SetBoundary(boundary)
}

// SetStart changes the compound object's "root"
func (w *Writer) SetStart(contentId string) error {
	cid, err := formatContentId(contentId)
	if err != nil {
		return err
	}

	w.start = cid
	return nil
}

// formatContentId parses given id and formats it as specified by
// RFC 5322
func formatContentId(contentId string) (string, error) {
	/*addr, err := mail.ParseAddress(contentId)
	if err != nil {
		return "", err
	}
	*/
	return contentId, nil
}

// SetType changes MIME mediaType of the compound object
func (w *Writer) SetType(mediaType string) error {
	if _, _, err := mime.ParseMediaType(mediaType); err != nil {
		return err
	}
	w.mediaType = mediaType

	return nil
}

// SetStartInfo changes startInfo of the compound object
func (w *Writer) SetStartInfo(info string) {
	w.startInfo = info
}

// FormDataContentType returns the Content-Type for a
// multipart/related with this Writer's Boundary, Start, Type and
// StartInfo.
func (w *Writer) FormDataContentType() string {
	params := map[string]string{
		"boundary": w.w.Boundary(),
	}

	if w.start != "" {
		params["start"] = w.start
	}
	if w.mediaType != "" {
		params["type"] = escapeQuotes(w.mediaType)
	}
	if w.startInfo != "" {
		params["start-info"] = escapeQuotes(w.startInfo)
	}

	return mime.FormatMediaType("multipart/related", params)
}

// CreateRoot creates a new multipart/related root section with the
// provided contentId, mediaType and header. The body of the root
// should be written to the returned Writer.
//
// header is used for adding additional information (e.g. Content-
// Transfer-Encoding), If Content-Id or Content-Type is specified in
// header, they will be overridden. If header is nil, creates a empty
// MIMEHeader.
func (w *Writer) CreateRoot(
	contentId string,
	mediaType string,
	header textproto.MIMEHeader,
) (io.Writer, error) {

	if w.rootPart {
		return nil, ErrRootExists
	}

	if header == nil {
		header = make(textproto.MIMEHeader)
	}

	if mediaType == "" {
		mediaType = DefaultMediaType
	}

	if err := w.SetType(mediaType); err != nil {
		return nil, err
	}
	header.Set("Content-Type", w.mediaType)
	w.rootMediaType = w.mediaType

	if contentId != "" {
		if err := w.SetStart(contentId); err != nil {
			return nil, err
		}
		header.Set("Content-ID", w.start)
	}

	w.firstPart = true
	w.rootPart = true

	return w.w.CreatePart(header)
}

// CreatePart is a wrapper around mulipart's Writer.CreatePart
func (w *Writer) CreatePart(
	contentId string,
	mediaType string,
	header textproto.MIMEHeader,
) (io.Writer, error) {

	var mediaType = DefaultMediaType

	if header == nil {
		header = make(textproto.MIMEHeader)
		header.Set("Content-Type", mediaType)
	} else if header.Get("Content-Type") != "" {
		mediaType = header.Get("Content-Type")
	}
	if mediaType == "" {
		mediaType = DefaultMediaType
	}

	if err := w.SetType(mediaType); err != nil {
		return nil, err
	}
	header.Set("Content-Type", w.mediaType)

	if contentId != "" {
		cid, err := formatContentId(contentId)
		if err != nil {
			return nil, err
		}
		header.Set("Content-ID", cid)
	}

	if w.firstPart == false {
		w.SetType(mediaType)
		//w.rootMediaType = w.mediaType
		w.firstPart = true
	}
	return w.w.CreatePart(header)
}

// Close is a wrapper around multipart's Writer.Close with additional errors.
func (w *Writer) Close() error {
	if w.mediaType != w.rootMediaType {
		return ErrTypeMatch
	}
	return w.w.Close()
}

// Helper func: escapeQuotes, borrowed from stdlib
// See http://golang.org/src/mime/multipart/writer.go#L115
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
