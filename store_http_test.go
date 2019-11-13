/* Copyright 2019 Kilobit Labs Inc. */

package stored // import "kilobit.ca/go/stored"

import "kilobit.ca/go/tested/assert"
import "testing"
import "strings"
import "net/http"
import "io/ioutil"
import "io"

func TestHttpStoreTest(t *testing.T) {
	assert.Expect(t, true, true)
}

type mockTransport struct{}

func (mt *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	code := http.StatusOK
	var r io.ReadCloser

	switch {
	case req.Method == "PUT":
		code = http.StatusCreated
	case req.Method == "GET" && req.URL.Path != "1":
		r = ioutil.NopCloser(strings.NewReader("Hello World!"))
	case req.Method == "GET" && req.URL.Path == "":
		r = ioutil.NopCloser(strings.NewReader("1,2,3"))
	case req.Method == "DELETE":
		code = http.StatusNoContent
	default:
		code = http.StatusMethodNotAllowed
	}

	return &http.Response{
		Status:     http.StatusText(code),
		StatusCode: code,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 0,
		Body:       r,
		Request:    req,
	}, nil
}

func TestNewHttpStore(t *testing.T) {

	hs := NewHttpStore(
		SimpleStoreReq("PUT", "/", AppendIDURLFunc),
		SimpleStoreReq("GET", "/", AppendIDURLFunc),
		SimpleStoreReq("GET", "/", AppendIDURLFunc),
		SimpleStoreReq("DELETE", "/", AppendIDURLFunc),
		StringMarshaler, StringUnmarshaler, StringIDUnmarshaler(","),
		OptUseTransport(&mockTransport{}),
	)

	err := hs.StoreItem("1", "Hello World!")
	if err != nil {
		t.Error(err)
	}

	var str string
	obj, err := hs.Retrieve("1", str)

	assert.Expect(t, "Hello World!", obj)
}
