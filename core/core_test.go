package core_test

import (
	"errors"
	"testing"

	"cloudNativeGo/core"
	"cloudNativeGo/transact"
)

func TestPut(t *testing.T) {
	tl, _ := transact.NewTransactionLogger("zero")
	store := core.NewKeyValueStore(tl)

	const key = "create-key"
	const value = "create-value"

	var val interface{}
	var contains bool

	defer delete(store.M, key)

	_, contains = store.M[key]

	if contains {
		t.Error("key/value already exists")
	}

	err := store.Put(key, value)

	if err != nil {
		t.Error(err)
	}

	val, contains = store.M[key]
	if !contains {
		t.Error("create failed")
	}

	if val != value {
		t.Error("val/value mismatch")
	}
}

func TestGet(t *testing.T) {
	tl, _ := transact.NewTransactionLogger("zero")
	store := core.NewKeyValueStore(tl)

	const key = "read-key"
	const value = "read-value"

	var val interface{}
	var err error

	defer delete(store.M, key)

	val, err = store.Get(key)
	if err == nil {
		t.Error("expected an error")
	}
	if !errors.Is(err, core.ErrorNoSuchKey) {
		t.Error("unexpected error: ", err)
	}

	store.M[key] = value

	val, err = store.Get(key)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if val != value {
		t.Error("val/value mismatch")
	}
}

func TestDelete(t *testing.T) {
	tl, _ := transact.NewTransactionLogger("zero")
	store := core.NewKeyValueStore(tl)

	const key = "delete-key"
	const value = "delete-value"

	var contains bool

	defer delete(store.M, key)

	store.M[key] = value

	_, contains = store.M[key]
	if !contains {
		t.Error("key/value doesn't exist")
	}

	_ = store.Delete(key)

	_, contains = store.M[key]
	if contains {
		t.Error("Delete failed")
	}
}
