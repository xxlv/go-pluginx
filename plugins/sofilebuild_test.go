package plugins

import (
	"os"
	"testing"

	msgchannel "github.com/xxlv/go-pluginx/build"
)

func TestCompileToSO(t *testing.T) {
	code := `
		package main
		
		import "fmt"
		
		func main() {
			fmt.Println("Hello, World!")
		}
	`

	soFilePath, err := CompileToSO(code, msgchannel.NilMessageChannel)
	defer os.Remove(soFilePath)
	if err != nil {
		t.Errorf("CompileToSO failed with error: %v", err)
	}

	if soFilePath == "" {
		t.Error("CompileToSO failed to return the .so file path")
	}

	invalidCode := `
		package main
		
		import "fmt"
		
		func main() {
			fmt.Println("Hello, World!"
		}
	`
	invalidSoFilePath, err := CompileToSO(invalidCode, msgchannel.NilMessageChannel)
	if err == nil {
		t.Error("CompileToSO should have returned an error for invalid code")
	}

	if invalidSoFilePath != "" {
		t.Error("CompileToSO should have returned an empty .so file path for invalid code")
	}

}
