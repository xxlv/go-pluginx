package build

import (
	"sync"
)

// Generate a channel based on the buildId
// The channel can receive any number of text messages
// The channel can be retrieved using the buildId
type MessageChannel chan string

var NilMessageChannel = make(MessageChannel)
var (
	mu       sync.RWMutex
	channels = make(map[string]MessageChannel)
)

func init() {
	// just remove NilMessageChannel messages
	go func() {
		for range NilMessageChannel {
		}
	}()

}

// NewChannel creates a new channel and associates it with the provided buildId.
func NewChannel(buildId string) MessageChannel {
	mu.Lock()
	defer mu.Unlock()

	ch := make(MessageChannel)
	channels[buildId] = ch
	return ch
}

// GetChannel retrieves the channel associated with the provided buildId.
func GetChannel(buildId string) (MessageChannel, bool) {
	mu.RLock()
	defer mu.RUnlock()

	ch, ok := channels[buildId]
	return ch, ok
}

// DeleteChannel deletes the channel associated with the provided buildId.
func DeleteChannel(buildId string) {
	mu.Lock()
	defer mu.Unlock()

	if ch, ok := channels[buildId]; ok {
		close(ch)
		delete(channels, buildId)
	}
}
