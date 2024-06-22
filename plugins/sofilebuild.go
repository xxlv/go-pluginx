package plugins

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"

	msgchannel "github.com/xxlv/go-pluginx/build"
)

// Share Object compile tool

// Compile any go code which in main pacakge
// if any output happend, mc channel will recive the logs
func CompileToSO(code string, mc msgchannel.MessageChannel) (string, error) {
	// Write function to a .go file
	tmpfile, err := os.CreateTemp("", "func.*.go")
	mc <- "Create empty temfile  success " + tmpfile.Name()
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(code)); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	// Generate hash of the function
	hash := sha256.Sum256([]byte(code))
	hashString := hex.EncodeToString(hash[:])
	mc <- "Gen Hash" + hashString

	soFilePath := filepath.Join(SoFileBuildPath, "func."+hashString+".so")
	// cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", soFilePath, tmpfile.Name())
	var cmd *exec.Cmd
	if os.Getenv("DEBUG") != "" {
		println("***************************************************************")
		println("prepare use debug ABI for creating new so file 🍺🍺🍺🍺🍺🍺🍺🍺🍺")
		println("***************************************************************")
		// in debug mode
		// this is very import for local debug,if not got error like below
		// `plugin was built with a different version of package internal/abi`
		cmd = exec.Command("go", "build", "-x", "-v", "-buildmode=plugin", "-gcflags", "all=-N -l", "-o", soFilePath, tmpfile.Name())

	} else {
		cmd = exec.Command("go", "build", "-x", "-v", "-buildmode=plugin", "-o", soFilePath, tmpfile.Name())

	}
	// Create stdout, stderr streams of type io.Reader
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	// Start command asynchronously
	err = cmd.Start()
	if err != nil {
		return "", err
	}

	// Create two channels to receive stdout and stderr
	out := make(chan string)
	outs := bufio.NewScanner(stdout)
	go func() {
		for outs.Scan() {
			out <- outs.Text()
		}
		close(out)
	}()

	errc := make(chan string)
	errs := bufio.NewScanner(stderr)
	go func() {
		for errs.Scan() {
			errc <- errs.Text()
		}
		close(errc)
	}()

	// Non-blockingly echo command output to message channel in go routine
	go func() {
		for {
			select {
			case line, ok := <-out:
				if ok && line != "" {
					mc <- line
				}
				if !ok {
					out = nil
				}
			case line, ok := <-errc:
				if ok && line != "" {
					mc <- line
				}
				if !ok {
					errc = nil
				}
			}
			if out == nil && errc == nil {
				return
			}
		}
	}()

	// Wait for command to finish
	err = cmd.Wait()
	if err != nil {
		return "", err
	}

	return soFilePath, nil
}
