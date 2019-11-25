StorEd
======

A set of simple storage interfaces implementing a Repository Pattern.

Status: In-Development

![StorEd System Components](doc/system_components.png)

In-Memory Storage:

```
	ms := NewMapStore()
	
	ms.StoreItem("1", "Hello World!")

	obj, _ := ms.Retrieve("1")
	
	s := obj.(string)

	assert.Expect(t, "Hello World!", s)
```

Custom REST storage client:

```
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
```

Custom REST storage service:
```
	ds := NewDataServer(
		"/test",
		stored.NewMapStore(),
		IncrIDGen(),
		OptSetEncoder("text/plain", PlainStringEncoder),
		OptSetDecoder("text/plain", PlainStringDecoder),
		OptSetLogger(log.New(ioutil.Discard, "", log.Ldate)),
	)
	
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

```

Features
--------

- Store data by only defining encoders and decoders.
- Non-opinionated interface for storing domain objects.
- Isolated domain and storage layers.
- Swap storage back-ends at will.
- Implement new storage implementations at will.
- Standard set of easy to set up components for In-Memory, WWW and soon DB Connectors.

Upcoming
- Simplified Http* interfaces for clients and servers.
- Implement a generic DB connector store.

Installation
------------

```
$ go get 'kilobit.ca/go/stored'
```

Building
--------

```
$ cd tested
$ go test -v
$ go build
```

Contribute
----------

Contributions and collaborations are welcome!

Please submit a pull request with any bug fixes or feature requests
that you have. All submissions imply consent to use / distribute under
the terms of the LICENSE.

Support
-------

Submit tickets through [github](https://github.com/kilobit/stored).

License
-------

See LICENSE.

--
Created: Nov 25, 2019
By: Christian Saunders <cps@kilobit.ca>
Copyright 2019 Kilobit Labs Inc.
