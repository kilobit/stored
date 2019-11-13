/* Copyright 2019 Kilobit Labs Inc. */

// A WWW Service implementation.
//
// Loosely inspired by https://blog.merovius.de/2017/06/18/how-not-to-use-an-http-router.html
//
package www // import "kilobit.ca/go/stored/www"

import "kilobit.ca/go/stored"
import "net/http"
import "strconv"
import "log"
import "os"
import "strings"
import "errors"
import "mime"
import "io"
import "path"
import "net/url"

type WWWOpt func(*DataServer)

type Encoder func(stored.Storable) ([]byte, error)

type Decoder func([]byte) (stored.Storable, error)

type DataServer struct {
	base     string
	encoders map[string]Encoder
	decoders map[string]Decoder
	store    stored.Store
	idgen    func(stored.Storable) string
	*log.Logger
}

func NewDataServer(base string, store stored.Store,
	idgen func(stored.Storable) string,
	opts ...WWWOpt) *DataServer {

	ds := &DataServer{
		base,
		map[string]Encoder{},
		map[string]Decoder{},
		store,
		idgen,
		log.New(os.Stderr, "www2: ", log.Ldate),
	}

	ds.Options(opts...)

	return ds
}

func (ds *DataServer) Options(opts ...WWWOpt) {
	for _, opt := range opts {
		opt(ds)
	}
}

// Shift a Path element from the URL.
//
func ShiftPath(p string) (head, tail string) {

	path := path.Clean(p)

	i := strings.Index(path[1:], "/")
	if i == -1 {
		i = len(path) - 1
	}

	i++

	head, err := url.QueryUnescape(path[1:i])
	if err != nil {
		head = path[1:i]
	}

	return head, path[i:]
}

// Currently, this function will return the first matched type in the
// accept header.
//
// Todo: handle accept parameters.
// Todo: Fix type selection to include prioritization etc.
//
func Acceptable(accept string, types map[string]Encoder) (string, Encoder) {
	for _, entry := range strings.Split(accept, ",") {
		mediatype, _, err := mime.ParseMediaType(entry)
		if err != nil {
			// handle mediatype parse error
			continue
		}

		enc, ok := types[mediatype]
		if ok {
			return mediatype, enc
		}
	}

	return "", nil
}

func (ds DataServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	req.URL.Path = strings.TrimPrefix(req.URL.Path, ds.base)

	switch req.Method {
	case "POST":
		ds.CreateData(res, req)
		return

	case "GET":
		ds.RetrData(res, req)
		return

	case "PUT":
		ds.UpdateData(res, req)
		return

	case "DELETE":
		ds.DeleteData(res, req)
		return

	default:
		ds.ServeError(
			http.StatusNotImplemented,
			"Invalid Method, "+req.Method+".",
			res, req)
	}
}

func (ds DataServer) ServeError(code int, msg string, res http.ResponseWriter, req *http.Request) {
	estr := req.Method + " " + req.URL.EscapedPath() + " " +
		http.StatusText(code) + " - " + msg
	ds.Println(estr)
	res.WriteHeader(code)
	res.Write([](byte)(estr))
}

func (ds DataServer) CreateData(res http.ResponseWriter, req *http.Request) {

	t := req.Header.Get("Content-Type")
	if t == "" {
		t = "application/octet-stream"
	}

	dec, ok := ds.decoders[t]
	if !ok {
		// Handle unknown content type.
		ds.ServeError(http.StatusUnsupportedMediaType,
			"Media type, "+t+"is not supported.",
			res, req)
		return
	}

	bs := make([]byte, req.ContentLength)
	_, err := req.Body.Read(bs)
	if err != nil && err != io.EOF {
		// handle read error.
		ds.ServeError(http.StatusBadRequest,
			"Could not read the request, "+err.Error(),
			res, req)
		return
	}

	obj, err := dec(bs)
	if err != nil {
		// handle decoding error
		ds.ServeError(http.StatusBadRequest,
			"Error decoding the request body, "+err.Error(),
			res, req)
		return
	}

	id := ds.idgen(obj)

	err = ds.store.StoreItem((stored.ID)(id), obj)
	if err != nil {
		// handle storage error
		ds.ServeError(http.StatusInternalServerError,
			"Error storing the object, "+err.Error(),
			res, req)
		return
	}

	res.Header().Add("Location", ds.base+"/"+id)
	res.WriteHeader(http.StatusCreated)
}

func (ds DataServer) RetrData(res http.ResponseWriter, req *http.Request) {

	id, _ := ShiftPath(req.URL.EscapedPath())
	if id == "" {
		// handle id error
		ds.ServeError(http.StatusBadRequest,
			"Invalid URL, "+req.URL.EscapedPath(),
			res, req)
		return
	}

	t, enc := Acceptable(req.Header.Get("Accept"), ds.encoders)
	if t == "" {
		// handle acceptable type error
		ds.ServeError(http.StatusNotAcceptable,
			"No acceptable response format is supported.",
			res, req)
		return
	}

	obj, err := ds.store.Retrieve((stored.ID)(id), "") // Probably a runtime error!!
	if err != nil {
		// handle storage error
		ds.ServeError(http.StatusNotFound,
			"The object was not found, "+err.Error(),
			res, req)
		return
	}

	bs, err := enc(obj)
	if err != nil {
		// handle encoding error
		ds.ServeError(http.StatusInternalServerError,
			"Failed to encode the store object, "+err.Error(),
			res, req)
		return
	}

	res.Header().Add("Content-Length", strconv.Itoa(len(bs)))
	res.WriteHeader(http.StatusOK)

	_, err = res.Write(bs)
	if err != nil {
		ds.Println(req.Method + " " + req.URL.EscapedPath() +
			"Write Error" + " - " + err.Error())
	}
}

func (ds DataServer) UpdateData(res http.ResponseWriter, req *http.Request) {

	id, _ := ShiftPath(req.URL.EscapedPath())
	if id == "" {
		// handle id error
		ds.ServeError(http.StatusBadRequest,
			"Invalid URL, "+req.URL.EscapedPath(),
			res, req)
		return
	}

	t := req.Header.Get("Content-Type")
	if t == "" {
		t = "application/octet-stream"
	}

	dec, ok := ds.decoders[t]
	if !ok {
		// Handle unknown content type.
		ds.ServeError(http.StatusUnsupportedMediaType,
			"Media type, "+t+" is not supported.",
			res, req)
		return
	}

	bs := make([]byte, req.ContentLength)
	_, err := req.Body.Read(bs)
	if err != nil && err != io.EOF {
		// handle read error.
		ds.ServeError(http.StatusBadRequest,
			"Could not read the request, "+err.Error(),
			res, req)
		return
	}

	obj, err := dec(bs)
	if err != nil {
		// handle decoding error
		ds.ServeError(http.StatusBadRequest,
			"Error decoding the request body, "+err.Error(),
			res, req)
		return
	}

	err = ds.store.StoreItem((stored.ID)(id), obj)
	if err != nil {
		// handle storage error
		ds.ServeError(http.StatusInternalServerError,
			"Error storing the object, "+err.Error(),
			res, req)
		return
	}

	res.WriteHeader(http.StatusNoContent)
}

func (ds DataServer) DeleteData(res http.ResponseWriter, req *http.Request) {

	id, _ := ShiftPath(req.URL.EscapedPath())
	if id == "" {
		// handle id error
		ds.ServeError(http.StatusBadRequest,
			"Invalid URL, "+req.URL.EscapedPath(),
			res, req)
		return
	}

	ds.store.Delete((stored.ID)(id))

	res.WriteHeader(http.StatusNoContent)
}

func PlainStringEncoder(s stored.Storable) ([]byte, error) {
	str, ok := s.(string)
	if !ok {
		// Handle type error
		return nil, errors.New("Object encodable as a string type.")
	}

	return ([]byte)(str), nil
}

func OptSetEncoder(t string, enc Encoder) WWWOpt {
	return func(ds *DataServer) {
		ds.encoders[t] = enc
	}
}

func OptSetDecoder(t string, dec Decoder) WWWOpt {
	return func(ds *DataServer) {
		ds.decoders[t] = dec
	}
}

func OptSetLogger(l *log.Logger) WWWOpt {
	return func(ds *DataServer) {
		ds.Logger = l
	}
}

func PlainStringDecoder(bs []byte) (stored.Storable, error) {
	return (string)(bs), nil
}

func IncrIDGen() func(stored.Storable) string {
	i := -1
	return func(obj stored.Storable) string {
		i++
		return strconv.Itoa(i)
	}
}
