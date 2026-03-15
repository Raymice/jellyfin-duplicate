package test

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func ReadJsonTestFile(t *testing.T, filePath string, v interface{}) {
	data, err := os.ReadFile(RootDir() + "/test/files/" + filePath)
	if err != nil {
		t.Fatalf("Failed to read input JSON: %v", err)
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
}

func ParseFromJsonFile[T any](t *testing.T, filePath string, target T) T {
	ReadJsonTestFile(t, filePath, &target)
	return target
}

func GetFuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	substrings := strings.Split(frame.Function, ".")
	return substrings[len(substrings)-1]
}

func GetTestUseCases(functionName string) []string {
	// list the diretories in the test/files/functionName directory and return the names of the directories as a list of strings
	files, err := os.ReadDir(RootDir() + "/test/files/" + functionName)
	if err != nil {
		panic(err)
	}
	var useCases []string
	for _, file := range files {
		if file.IsDir() {
			useCases = append(useCases, file.Name())
		}
	}
	return useCases
}
