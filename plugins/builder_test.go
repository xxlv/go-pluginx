package plugins

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/xxlv/go-pluginx/storage/adapters"
)

func TestPluginBuilder_Build(t *testing.T) {
	pb := &PluginBuilder{
		Storage: &storages.MemoryStorage{
			Kv: make(map[string]string),
		},
	}
	bp := &BuildPayload{
		PluginFile: "",
		Code:       "",
	}

	// Test case where neither SoFile nor Code is provided
	br, err := pb.Build(context.Background(), bp, nil)
	assert.NotNil(t, err, "Expected error when no SoFile or Code is provided")
	assert.Equal(t, "", br.BuildId, "Expected BuildId to be empty when no SoFile or Code is provided")

}

func TestPluginBuilder_SetBuildStatus(t *testing.T) {
	// Create a mock PluginBuilder with a mock storage implementation
	mockStorage := &storages.MemoryStorage{} // create a mock storage implementation, or use a testing framework with mocks
	pb := &PluginBuilder{Storage: mockStorage, Logger: slog.Default()}

	// Call setBuildStatus with some test data
	buildID := "testBuildID"
	status := Running // replace with the desired status
	pb.setBuildStatus(context.Background(), buildID, status)

	// Retrieve the stored build info
	key := BuildInfoPrefix + buildID
	storedData, err := mockStorage.Get(key)
	if storedData == "" || err != nil {
		t.Fatalf("Error getting build info: %v", err)
	}

	t.Log(storedData)

	// Unmarshal the stored data
	var storedResult []*BuildResultInfo
	err = json.Unmarshal([]byte(storedData), &storedResult)
	if err != nil {
		t.Fatalf("Error unmarshalling build info: %v", err)
	}

	// Perform assertions based on the test case
	if len(storedResult) != 1 {
		t.Fatalf("Expected one BuildResultInfo, got %d", len(storedResult))
	}

}
