StorEd
======

A simple storage Interface for Golang.

The idea is to decouple domain logic from the storage logic without
hiding or prescribing the details of either.

Consider building a set of data objects and defining encoding /
decoding schemes for these objects.  Now with the same Store
interface, these data objects can be kept in-memory, on the local
disk, on a remote REST service, in a database or in a cloud storage
system like Firebase or Dynamo.

When your needs change, simply reconfigure the storage layer for a new
destination without making any changes in the domain.

As the need arises for a custom storage or encoding / decoding
mechanism, there is no limitation as none of the workings are hidden
and custom components can be implemented without loosing the benefits
otherwise inherent in the system.

Note that this flexibility leaves the onus of understanding the
implications of a particular implementation on the developer.
Developers should understand the encoding formats and transport
mechanisms involved.

Status: In-Development

Please help me to continue the design and development of this system.

This is a work in progress and while the current implementation is
useful, many rough edges needs smoothing and the API would benefit
from polishing.

Consider this example system layout:

![StorEd System Components](doc/system_components.png){ style="width: 100%" }

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

- Store data objects by defining encoders and decoders or using
  standard libraries.
- Isolate domain and storage layers.
- Swap storage back-ends at will.
- Implement new storage connectors or back-ends at will.
- Use the Go standard library interface definitions to add middleware
  for authentication, encoders etc.
- Use dependency injection to customize the behaviour of the stores.
- Leverage a standard set of easy to set up components for In-Memory,
  WWW and soon DB Connectors.

Upcoming:

- Simplified Http* DI interfaces for clients and servers.
- Implement a generic DB connector store.
- Mix and match storage implementations.

Issues
------

- The current method for defining HTTP headers is cumbersome.
- Some naming conventions, particularly for standard HTTP DI
  components is unclear.

Installation
------------

```
$ go get 'kilobit.ca/go/stored'
```

Building
--------

```
$ cd stored
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
