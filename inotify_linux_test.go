// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package inotify

import (
	"io/ioutil"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func TestInotifyEvents(t *testing.T) {
	// Create an inotify watcher instance and initialize it
	watcher, err := NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher failed: %s", err)
	}

	dir, err := ioutil.TempDir("", "inotify")
	if err != nil {
		t.Fatalf("TempDir failed: %s", err)
	}
	defer os.RemoveAll(dir)

	// Add a watch for "_test"
	err = watcher.Watch(dir)
	if err != nil {
		t.Fatalf("Watch failed: %s", err)
	}

	// Receive errors on the error channel on a separate goroutine
	go func() {
		for err := range watcher.Error {
			t.Fatalf("error received: %s", err)
		}
	}()

	testFile := dir + "/TestInotifyEvents.testfile"

	// Receive events on the event channel on a separate goroutine
	eventstream := watcher.Event
	var eventsReceived int32 = 0
	done := make(chan bool)
	go func() {
		for event := range eventstream {
			atomic.AddInt32(&eventsReceived, 1)
			t.Logf("event received: %s", event)
		}
		done <- true
	}()

	// Create a file
	// This should add IN_CREATE and IN_OPEN events
	f, err := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("creating test file: %s", err)
	}
	// This should add an IN_CLOSE event
	err = f.Close()
	if err != nil {
		t.Fatalf("closing test file: %s", err)
	}
	// This should add an IN_DELETE event
	err = os.Remove(testFile)
	if err != nil {
		t.Fatalf("removing test file: %s", err)
	}
	// This should add IN_DELETE_SELF and IN_IGNORED events
	err = os.Remove(dir)
	if err != nil {
		t.Fatalf("removing test dir: %s", err)
	}
	// We expect this event to be received almost immediately, but let's wait 100 ms to be sure
	time.Sleep(100 * time.Millisecond)
	received := atomic.AddInt32(&eventsReceived, 0)
	if received == 0 {
		t.Fatal("inotify event hasn't been received after 100 milliseconds")
	}
	if received != 6 {
		t.Fatal("expected 6 inotify events got", received)
	}

	// Try closing the inotify instance
	t.Log("calling Close()")
	watcher.Close()
	t.Log("waiting for the event channel to become closed...")
	select {
	case <-done:
		t.Log("event channel closed")
	case <-time.After(1 * time.Second):
		t.Fatal("event stream was not closed after 1 second")
	}
	if len(watcher.watches) > 0 || len(watcher.paths) > 0 {
		t.Error("Watches were not properly removed")
	}
}

func TestInotifyClose(t *testing.T) {
	watcher, _ := NewWatcher()
	watcher.Close()

	done := make(chan bool)
	go func() {
		watcher.Close()
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("double Close() test failed: second Close() call didn't return")
	}

	err := watcher.Watch(os.TempDir())
	if err == nil {
		t.Fatal("expected error on Watch() after Close(), got nil")
	}
}
