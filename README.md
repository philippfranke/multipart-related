# mRelated

mRelated implements mime-type multipart/related [RFC 2387](http://tools.ietf.org/html/rfc2387)

**Documentation:** [![GoDoc](https://godoc.org/github.com/philippfranke/mrelated?status.svg)](https://godoc.org/github.com/philippfranke/mrelated)

**Build Status:** [![Build Status](https://travis-ci.org/philippfranke/mrelated.svg?branch=master)](https://travis-ci.org/philippfranke/mrelated)

go-github requires Go version 1.1 or greater.

## Usage
```go
import "github.com/philippfranke/mrelated"
```

```go
reader := mrelated.NewReader(b, boundary)
aggregate, _ := reader.ReadAggregate()
for _, obj :=  range aggregate.Object {
  fmt.Printf("Content: %s", obj.String())
}

```

## License

This library is distributed under the BSD-style license found in the [LICENSE](./LICENSE)
file.

