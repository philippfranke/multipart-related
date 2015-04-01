# multipart-related

multipart-related implements MIME multipart/related in go, [RFC 2387](http://tools.ietf.org/html/rfc2387)

**Documentation:** [![GoDoc](https://godoc.org/github.com/philippfranke/multipart-related/related?status.svg)](https://godoc.org/github.com/philippfranke/multipart-related/related)

**Build Status:** [![Build Status](https://travis-ci.org/philippfranke/multipart-related.svg?branch=master)](https://travis-ci.org/philippfranke/multipart-related)

**Test Coverage:** [![Coverage Status](https://coveralls.io/repos/philippfranke/multipart-related/badge.svg)](https://coveralls.io/r/philippfranke/multipart-related)

multipart-related requires Go version 1.2 or greater.

## What is multipart-related
The Package related implements MIME multipart/related parsing, as defined in [RFC 2387](http://tools.ietf.org/html/rfc2387).

*See [Wikipedia](http://en.wikipedia.org/wiki/MIME#Related):*
>A multipart/related is used to indicate that each message part is a component of an aggregate whole. It is for compound objects consisting of several inter-related components - proper display cannot be achieved by individually displaying the constituent parts. The message consists of a root part (by default, the first) which reference other parts inline, which may in turn reference other parts. Message parts are commonly referenced by the "Content-ID" part header. The syntax of a reference is unspecified and is instead dictated by the encoding or protocol used in the part.

>One common usage of this subtype is to send a web page complete with images in a single message. The root part would contain the HTML document, and use image tags to reference images stored in the latter parts.

*Compatible with Google's [Drive REST API](https://developers.google.com/drive/web/manage-uploads)*

## Usage
```go
import "github.com/philippfranke/multipart-related/related"
```
### Writer

```go
content := []byte(`Life? Don't talk to me about life!`)   // Douglas Adams, The Hitchhiker's Guide to the Galaxy
var b bytes.Buffer
w := related.NewWriter(&b)

rootPart, err := w.CreateRoot("m@rv.in", "H2/G2", nil)
if err != nil {
  panic(err)
}

rootPart.Write(content[:5])

nextPart, err := w.CreatePart("", nil)
if err != nil {
  panic(err)
}
nextPart.Write(content[5:])

if err := w.Close(); err != nil {
  panic(err)
}

fmt.Printf("The compound Object Content-Type:\n %s \n", w.FormDataContentType())
fmt.Fprintf(os.Stdout, "Body: \n %s", b.String())
```

### Reader

```go
header := `multipart/related; boundary=example-1;
                start="<b@c.d>";
                type="a/b"`

msg := strings.NewReader(`--example-1
Content-Type: a/b
Content-ID: <a@b.c>

Life?
--example-1
Content-Type: b/c
Content-Transfer-Encoding: base64
Content-ID: <b@c.d>

RG9uJ3QgdGFsayB0byBtZSBhYm91dCBsaWZlIQ==

--example-1--`)

  mediaType, params, err := mime.ParseMediaType(header)
  if err != nil {
    panic(err)
  }
  if mediaType != "multipart/related" {
    panic(nil)
  }

  r := related.NewReader(msg, params)
  object, err := r.ReadObject()
  if err != nil {
    panic(err)
  }
  for i, part := range object.Values {
    b, err := ioutil.ReadAll(part)
    if err != nil {
      panic(err)
    }
    fmt.Printf("%d.: %s \n", i, string(b))
  }
```

## License

This library is distributed under the BSD-style license found in the [LICENSE](./LICENSE)
file.
