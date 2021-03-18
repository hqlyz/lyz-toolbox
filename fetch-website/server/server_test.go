package server

import (
	"testing"
)

func TestQueueTask(t *testing.T) {
	server := New(true)
	server.Enqueue("one")
	server.Enqueue("two")
	server.Enqueue("three")
	expected := []string{"one", "two", "three"}
	if !testSlice(server.taskList, expected) {
		t.Fatalf("expected: %v, got: %v\n", expected, server.taskList)
	}
	server.dequeue()
	expected = []string{"two", "three"}
	if !testSlice(server.taskList, expected) {
		t.Fatalf("expected: %v, got: %v\n", expected, server.taskList)
	}
	server.dequeue()
	expected = []string{"three"}
	if !testSlice(server.taskList, expected) {
		t.Fatalf("expected: %v, got: %v\n", expected, server.taskList)
	}
	server.dequeue()
	expected = []string{}
	if !testSlice(server.taskList, expected) {
		t.Fatalf("expected: %v, got: %v\n", expected, server.taskList)
	}
	server.dequeue()
	expected = []string{}
	if !testSlice(server.taskList, expected) {
		t.Fatalf("expected: %v, got: %v\n", expected, server.taskList)
	}
}

func testSlice(s []string, expected []string) bool {
	if len(s) != len(expected) {
		return false
	}

	for k, v := range s {
		if v != expected[k] {
			return false
		}
	}

	return true
}
