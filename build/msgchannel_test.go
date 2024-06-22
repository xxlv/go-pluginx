package build

import (
	"testing"
)

func TestChannelOperations(t *testing.T) {
	buildId := "test-build"

	// Test NewChannel
	ch := NewChannel(buildId)
	if ch == nil {
		t.Fatal("NewChannel returned nil")
	}

	// Test GetChannel
	retrievedCh, ok := GetChannel(buildId)
	if !ok {
		t.Fatal("GetChannel did not find the channel")
	}

	if retrievedCh != ch {
		t.Fatal("GetChannel returned a different channel")
	}

	// Test DeleteChannel
	DeleteChannel(buildId)
	_, ok = GetChannel(buildId)
	if ok {
		t.Fatal("GetChannel found the channel after it was deleted")
	}
}

func TestChannelCommunication(t *testing.T) {
	buildId := "test-build"
	message := "test message"

	ch := NewChannel(buildId)
	go func() {
		ch <- message
	}()

	retrievedMessage := <-ch
	if retrievedMessage != message {
		t.Fatalf("Expected %s, got %s", message, retrievedMessage)
	}

	DeleteChannel(buildId)
}
