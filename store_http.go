/* Copyright 2019 Kilobit Labs Inc. */

package stored // import "kilobit.ca/go/stored"

import "bufio"
import "bytes"
import "encoding/json"
import "io"
import "io/ioutil"
import "net/http"
import "net/url"
import "strings"
import "time"

type HttpStoreError string

func (err HttpStoreError) Error() string {
	return (string)(err)
}

func NewHttpStoreError(msg string) HttpStoreError {
	return (HttpStoreError)(msg)
}

type HttpStoreOpt func(*HttpStore)

func OptUseClient(c *http.Client) HttpStoreOpt {
	return func(h *HttpStore) {
		h.c = c
	}
}

func OptUseTransport(t http.RoundTripper) HttpStoreOpt {
	return func(h *HttpStore) {
		h.c.Transport = t
	}
}

type HttpStoreReq func(ID) (*http.Request, error)
type HttpStoreMarshaler func(Storable) (io.ReadCloser, int64, error)
type HttpStoreUnmarshaler func(io.Reader) (Storable, error)
type HttpStoreIDUnmarshaler func(io.Reader) ([]ID, error)

// A HTTP Client based store.
//
// Note: This store will require a remote service to connect to.
type HttpStore struct {
	c            *http.Client           // The HTTP Client
	storeRequest HttpStoreReq           // StoreItem request
	retrRequest  HttpStoreReq           // Retrieve request
	listRequest  HttpStoreReq           // List request
	delRequest   HttpStoreReq           // Delete request
	marshal      HttpStoreMarshaler     // Object marshaler
	unmarshal    HttpStoreUnmarshaler   // Object unmarshaler
	unmarshalIDs HttpStoreIDUnmarshaler // ID list unmarshaler
}

func NewHttpStore(sr, rr, lr, dr HttpStoreReq,
	m HttpStoreMarshaler, u HttpStoreUnmarshaler,
	ui HttpStoreIDUnmarshaler,
	opts ...HttpStoreOpt) *HttpStore {

	s := &HttpStore{
		&http.Client{
			Timeout: time.Second * 10,
		},
		sr, rr, lr, dr, m, u, ui,
	}

	s.Options(opts...)

	return s
}

func (s *HttpStore) Options(opts ...HttpStoreOpt) {
	for _, opt := range opts {
		opt(s)
	}
}

func (s *HttpStore) StoreItem(id ID, obj Storable) error {

	req, err := s.storeRequest(id)
	if err != nil {
		return err
	}

	req.Body, req.ContentLength, err = s.marshal(obj)

	res, err := s.c.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		return NewHttpStoreError("Failed response from server: " + http.StatusText(res.StatusCode))
	}

	return nil
}

func (s *HttpStore) Retrieve(id ID, dst Storable) (Storable, error) {

	req, err := s.retrRequest(id)
	if err != nil {
		return nil, err
	}

	res, err := s.c.Do(req)
	if err != nil {
		return nil, err
	}

	dst, err = s.unmarshal(res.Body)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

// TODO: Handle Errors
func (s *HttpStore) List() []ID {

	req, err := s.listRequest("")
	if err != nil {
		return []ID{}
	}

	res, err := s.c.Do(req)
	if err != nil {
		return []ID{}
	}

	ids, err := s.unmarshalIDs(res.Body)
	if err != nil {
		return []ID{}
	}

	return ids
}

func (s *HttpStore) Apply(f ItemHandler, dst Storable) error {

	ids := s.List()
	for _, id := range ids {

		dst, err := s.Retrieve(id, dst)
		if err != nil {
			return err
		}

		err = f(id, dst)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Handle errors
func (s *HttpStore) Delete(id ID) {
	req, _ := s.delRequest(id)
	s.c.Do(req)
}

type URLFunc func(base string, id ID) (*url.URL, error)

func AppendIDURLFunc(base string, id ID) (*url.URL, error) {

	return url.Parse(base + "/" + (string)(id))
}

func SimpleStoreReq(method, base string, urlf URLFunc, hdrs *http.Header) HttpStoreReq {
	return func(id ID) (*http.Request, error) {

		url, err := urlf(base, id)
		if err != nil {
			return nil, err
		}

		return &http.Request{
			Method: method,
			URL:    url,
			Header: *hdrs,
		}, nil
	}
}

func StringMarshaler(obj Storable) (io.ReadCloser, int64, error) {
	str, ok := obj.(string)
	if !ok {
		return nil, 0, NewHttpStoreError("Expected the Storable to be a string.")
	}

	return ioutil.NopCloser(strings.NewReader(str)), (int64)(len(str)), nil
}

func StringUnmarshaler(r io.Reader) (Storable, error) {
	s := bufio.NewScanner(r)
	if !s.Scan() {
		return nil, NewHttpStoreError("Error scanning the string.")
	}

	return s.Text(), nil
}

func StringIDUnmarshaler(sep string) HttpStoreIDUnmarshaler {
	return func(r io.Reader) ([]ID, error) {
		bs := []byte{}
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}

		idstrs := strings.Split((string)(bs), sep)
		ids := make([]ID, len(idstrs))
		for i := range idstrs {
			ids[i] = (ID)(idstrs[i])
		}

		return ids, nil
	}
}

func JSONMarshaler(obj Storable) (io.Reader, error) {
	bs, err := json.Marshal(obj)
	return bytes.NewReader(bs), err
}

func JSONUnmarshaler(obj Storable) HttpStoreUnmarshaler {
	return func(r io.Reader) (Storable, error) {
		//		dec := json.NewDecoder(r)
		return nil, nil

	}
}
