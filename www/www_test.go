/* Copyright 2019 Kilobit Labs Inc. */

// Tests for the WWW2 Service implementation.
//
package www // import "kilobit.ca/go/stored/www"

import "kilobit.ca/go/stored"
import "kilobit.ca/go/tested/assert"
import "testing"
import "net/http/httptest"
import "strings"
import "net/http"
import "io/ioutil"
import "log"

func TestWWW2Test(t *testing.T) {
	assert.Expect(t, true, true)
}

var ds *DataServer = NewDataServer(
	"/test",
	stored.NewMapStore(),
	IncrIDGen(),
	OptSetEncoder("text/plain", PlainStringEncoder),
	OptSetDecoder("text/plain", PlainStringDecoder),
	OptSetLogger(log.New(ioutil.Discard, "", log.Ldate)),
)

func TestWWW2New(t *testing.T) {

	req := httptest.NewRequest(
		"POST",
		"/test/",
		strings.NewReader("Hello World!"),
	)
	req.Header.Add("Content-Type", "text/plain")

	res := httptest.NewRecorder()
	ds.ServeHTTP(res, req)
}

func TestWWW2InvalidMethod(t *testing.T) {

	req := httptest.NewRequest(
		"FOO",
		"/test/",
		strings.NewReader("Hello World!"),
	)
	req.Header.Add("Content-Type", "text/plain")

	res := httptest.NewRecorder()
	ds.ServeHTTP(res, req)
	assert.Expect(t, http.StatusNotImplemented, res.Code)
}

func TestWWW2CreateAndRetrieve(t *testing.T) {

	req := httptest.NewRequest(
		"POST",
		"/test/",
		strings.NewReader("Hello World!"),
	)
	req.Header.Add("Content-Type", "text/plain")

	res := httptest.NewRecorder()
	ds.ServeHTTP(res, req)
	assert.Expect(t, http.StatusCreated, res.Code)

	loc := res.Header().Get("Location")
	if loc == "" {
		t.Error("Location not set.")
	}

	req = httptest.NewRequest("GET", loc, nil)
	req.Header.Add("Accept", "text/plain")
	res = httptest.NewRecorder()
	ds.ServeHTTP(res, req)

	assert.Expect(t, http.StatusOK, res.Code)

	bs := make([]byte, res.Result().ContentLength)
	i, err := res.Result().Body.Read(bs)
	if err != nil {
		t.Error(err)
	}

	assert.Expect(t, len("Hello World!"), i)
	assert.Expect(t, "Hello World!", (string)(bs))
}

func TestWWW2CreateAndUpdate(t *testing.T) {

	req := httptest.NewRequest(
		"POST",
		"/test/",
		strings.NewReader("Hello World!"),
	)
	req.Header.Add("Content-Type", "text/plain")

	res := httptest.NewRecorder()
	ds.ServeHTTP(res, req)
	assert.Expect(t, http.StatusCreated, res.Code)

	loc := res.Header().Get("Location")
	if loc == "" {
		t.Error("Location not set.")
	}

	req = httptest.NewRequest(
		"PUT",
		loc,
		strings.NewReader("Hello!"),
	)
	req.Header.Add("Content-Type", "text/plain")
	res = httptest.NewRecorder()
	ds.ServeHTTP(res, req)

	assert.Expect(t, http.StatusNoContent, res.Code)

	req = httptest.NewRequest("GET", loc, nil)
	req.Header.Add("Accept", "text/plain")
	res = httptest.NewRecorder()
	ds.ServeHTTP(res, req)

	assert.Expect(t, http.StatusOK, res.Code)

	bs := make([]byte, res.Result().ContentLength)
	i, err := res.Result().Body.Read(bs)
	if err != nil {
		t.Error(err)
	}

	assert.Expect(t, len("Hello!"), i)
	assert.Expect(t, "Hello!", (string)(bs))
}

func TestWWW2CreateAndDelete(t *testing.T) {

	req := httptest.NewRequest(
		"POST",
		"/test/",
		strings.NewReader("Hello World!"),
	)
	req.Header.Add("Content-Type", "text/plain")

	res := httptest.NewRecorder()
	ds.ServeHTTP(res, req)
	assert.Expect(t, http.StatusCreated, res.Code)

	loc := res.Header().Get("Location")
	if loc == "" {
		t.Error("Location not set.")
	}

	req = httptest.NewRequest(
		"DELETE",
		loc,
		nil,
	)
	res = httptest.NewRecorder()
	ds.ServeHTTP(res, req)

	assert.Expect(t, http.StatusNoContent, res.Code)

	req = httptest.NewRequest("GET", loc, nil)
	req.Header.Add("Accept", "text/plain")
	res = httptest.NewRecorder()
	ds.ServeHTTP(res, req)

	assert.Expect(t, http.StatusNotFound, res.Code)
}

func TestWWW2HttpStoreClient(t *testing.T) {

	ds := NewDataServer(
		"/",
		stored.NewMapStore(),
		IncrIDGen(),
		OptSetEncoder("text/plain", PlainStringEncoder),
		OptSetDecoder("text/plain", PlainStringDecoder),
		OptSetLogger(log.New(ioutil.Discard, "", log.Ldate)),
	)

	srv := httptest.NewServer(ds)

	hdrs := &http.Header{}
	hdrs.Add("Accept", "text/plain")
	hdrs.Add("Content-Type", "text/plain")
	
	hs := stored.NewHttpStore(
		stored.SimpleStoreReq("PUT", srv.URL + "/", stored.AppendIDURLFunc, hdrs),
		stored.SimpleStoreReq("GET", srv.URL + "/", stored.AppendIDURLFunc, hdrs),
		stored.SimpleStoreReq("GET", srv.URL + "/", stored.AppendIDURLFunc, hdrs),
		stored.SimpleStoreReq("DELETE", srv.URL + "/", stored.AppendIDURLFunc, hdrs),
		stored.StringMarshaler, stored.StringUnmarshaler,
		stored.StringIDUnmarshaler(","),
		stored.OptUseClient(srv.Client()),
	)
	
	err := hs.StoreItem("1", "Hello World!")
	if err != nil {
		t.Error(err)
	}

	obj, err := hs.Retrieve("1", "")
	if err != nil {
		t.Error(err)
	}

	s, ok := obj.(string)
	if !ok {
		t.Errorf("Retrieved object is not a string.")
	}
	
	assert.Expect(t, "Hello World!", s)
}
