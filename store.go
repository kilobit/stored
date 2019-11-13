/* Copyright 2019 Kilobit Labs Inc. */

package stored // import "kilobit.ca/go/stored"

type StoreError string

func NewStoreError(msg string) StoreError {
	return (StoreError)(msg)
}

func (err StoreError) Error() string {
	return (string)(err)
}

type ID string

type Storable interface{}

type ItemHandler func(ID, Storable) error

// Interface defining a generic repository pattern for data access.
//
// TODO: Add criteria as a parameter to filter selected results.
//
type Store interface {
	StoreItem(ID, Storable) error
	Retrieve(ID, Storable) (Storable, error)
	List() []ID
	Apply(ItemHandler, Storable) error
	Delete(ID)
}

// An in-memory map based store.
//
// Note: This store is volatile and disapears on application exit.
type MapStore map[ID]Storable

func NewMapStore() MapStore {
	return make(MapStore)
}

func (s MapStore) StoreItem(id ID, obj Storable) error {

	s[id] = obj

	return nil
}

func (s MapStore) Retrieve(id ID, dst Storable) (Storable, error) {

	dst, ok := s[id]
	if !ok {
		return nil, NewStoreError("Store object not found.")
	}

	return dst, nil
}

func (s MapStore) List() []ID {

	ids := []ID{}

	for id := range s {
		ids = append(ids, id)
	}

	return ids
}

func (s MapStore) Apply(f ItemHandler, dst Storable) error {
	for id, dst := range s {
		err := f(id, dst)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s MapStore) Delete(id ID) {
	delete(s, id)
}
