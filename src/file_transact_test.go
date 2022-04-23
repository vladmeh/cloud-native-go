package main

import (
	"os"
	"testing"
)

func fileExist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func TestCreateLogger(t *testing.T) {
	const filename = "create-logger.txt"
	defer func() {
		err := os.Remove(filename)
		if err != nil {
			t.Error(err)
		}
	}()

	tl, err := NewFileTransactionLogger(filename)
	defer func(tl *FileTransactionLogger) {
		_ = tl.Close()
	}(tl)

	if tl == nil {
		t.Error("Logger is null")
	}

	if err != nil {
		t.Errorf("Got error: %s", err)
	}

	if !fileExist(filename) {
		t.Errorf("File %s in not exist", filename)
	}
}

func TestWriteAppend(t *testing.T) {
	const filename = "write-append.txt"
	defer func() {
		err := os.Remove(filename)
		if err != nil {
			t.Error(err)
		}
	}()

	tl, err := NewFileTransactionLogger(filename)
	if err != nil {
		t.Error(err)
	}

	tl.Run()
	defer func(tl *FileTransactionLogger) {
		_ = tl.Close()
	}(tl)

	chev, cherr := tl.ReadEvents()
	for e := range chev {
		t.Log(e)
	}
	err = <-cherr
	if err != nil {
		t.Error(err)
	}

	tl.WritePut("my-key", "my-value")
	tl.WritePut("my-key", "my-value")
	tl.Wait()

	if tl.lastSequence != 2 {
		t.Errorf("Last sequence mismatch (expected 2; got %d)", tl.lastSequence)
	}

	tl2, err := NewFileTransactionLogger(filename)
	if err != nil {
		t.Error(err)
	}

	tl2.Run()
	defer func(tl2 *FileTransactionLogger) {
		_ = tl2.Close()
	}(tl2)

	chev, cherr = tl2.ReadEvents()
	for e := range chev {
		t.Log(e)
	}
	err = <-cherr
	if err != nil {
		t.Error(err)
	}

	tl2.WritePut("my-key", "my-value3")
	tl2.WritePut("my-key2", "my-value4")
	tl2.Wait()

	if tl2.lastSequence != 4 {
		t.Errorf("Last sequence mismatch (expected 4; got %d)", tl2.lastSequence)
	}
}

func TestWritePut(t *testing.T) {
	const filename = "write-put.txt"
	defer func() {
		_ = os.Remove(filename)
	}()

	tl, _ := NewFileTransactionLogger(filename)
	tl.Run()

	defer func(tl *FileTransactionLogger) {
		_ = tl.Close()
	}(tl)

	tl.WritePut("key", "value")
	tl.WritePut("key1", "value1")
	tl.WritePut("my key-2", "value 2")
	tl.WritePut("my key 3", "value 3")
	tl.Wait()

	tl2, _ := NewFileTransactionLogger(filename)
	chev, cherr := tl2.ReadEvents()
	defer func(tl2 *FileTransactionLogger) {
		_ = tl2.Close()
	}(tl2)

	for e := range chev {
		t.Log(e)
	}

	err := <-cherr
	if err != nil {
		t.Error(err)
	}

	if tl.lastSequence != tl2.lastSequence {
		t.Errorf("Last sequence mismatch (%d vs. %d)", tl.lastSequence, tl2.lastSequence)
	}
}
