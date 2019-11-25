/* Copyright 2019 Kilobit Labs Inc. */

package stored // import "kilobit.ca/go/stored"

import "kilobit.ca/go/tested/assert"
import "testing"

func TestStoreTest(t *testing.T) {
	assert.Expect(t, true, true)
}

func TestNewMapStore(t *testing.T) {

	ms := NewMapStore()
	
	err := ms.StoreItem("1", "Hello World!")
	if err != nil {
		t.Error(err)
	}

	obj, err := ms.Retrieve("1")
	if err != nil {
		t.Error(err)
	}
	
	s, ok := obj.(string)
	if !ok {
		t.Errorf("Returned object was not a string.")
	}

	assert.Expect(t, "Hello World!", s)
}
