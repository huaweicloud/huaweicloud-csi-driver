package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"
)

func TestAppendToFile(t *testing.T) {
	ex, path := createFile(t, "appendToFile.txt")

	tests := []struct {
		name        string
		path        string
		content     string
		expected    bool
		description string
	}{
		{
			name:        "test1",
			path:        filepath.Join(filepath.Dir(ex), "test1.txt"),
			content:     "test1 content",
			expected:    false,
			description: "the file does not exist",
		},
		{
			name:        "test2",
			path:        path,
			content:     "test2 content",
			expected:    true,
			description: "the file already exists",
		},
		{
			name:        "test3",
			path:        "xxx/D81u_7cds12e.txt",
			content:     "test3 content",
			expected:    false,
			description: "the path does not exist",
		},
		{
			name:        "test4",
			path:        "",
			content:     "test4 content",
			expected:    false,
			description: "path is empty",
		},
	}

	// test function.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := AppendToFile(testCase.path, testCase.content)
			if testCase.expected == true && err != nil {
				t.Errorf("expected: %v, got: %v ", testCase.expected, err)
			}
			if testCase.expected == false && err == nil {
				t.Errorf("expected error but get nil")
			}

			// check the content of last line
			if testCase.expected {
				err = checkFileContent(testCase.path, testCase.content, t)
				if err != nil {
					t.Errorf("check file content failed, the AppendToFile() failed to append content to file: %v", err)
				}
			}
		})
	}
}

func TestWriteToFile(t *testing.T) {
	ex, path := createFile(t, "writeToFile.txt")

	tests := []struct {
		name        string
		path        string
		content     string
		expected    bool
		description string
	}{
		{
			name:        "test1",
			path:        filepath.Join(filepath.Dir(ex), "test1.txt"),
			content:     "test1 content",
			expected:    true,
			description: "the file does not exist",
		},
		{
			name:        "test2",
			path:        path,
			content:     "test2 content",
			expected:    true,
			description: "the file already exists",
		},
		{
			name:        "test3",
			path:        "xxx/D81u_7cds12e.txt",
			content:     "test3 content",
			expected:    false,
			description: "the path does not exist",
		},
		{
			name:        "test4",
			path:        "",
			content:     "test4 content",
			expected:    false,
			description: "path is empty",
		},
	}

	// test function.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := WriteToFile(testCase.path, testCase.content)
			if testCase.expected == true && err != nil {
				t.Errorf("expected: %v, got: %v ", testCase.expected, err)
			}
			if testCase.expected == false && err == nil {
				t.Errorf("expected error but get nil")
			}

			// check the content of last line
			if testCase.expected {
				err = checkFileContent(testCase.path, testCase.content, t)
				if err != nil {
					t.Errorf("check file content failed, the WriteToFile() failed to append content to file: %v", err)
				}
			}
		})
	}
}

func TestDeleteFile(t *testing.T) {
	ex, path := createFile(t, "deleteFile.txt")

	tests := []struct {
		name        string
		path        string
		expected    bool
		description string
	}{
		{
			name:        "test1",
			path:        filepath.Join(filepath.Dir(ex), "test1.txt"),
			expected:    false,
			description: "the file does not exist",
		},
		{
			name:        "test2",
			path:        path,
			expected:    true,
			description: "the file already exists",
		},
		{
			name:        "test3",
			path:        "xxx/D81u_7cds12e.txt",
			expected:    false,
			description: "the path does not exist",
		},
		{
			name:        "test4",
			path:        "",
			expected:    false,
			description: "path is empty",
		},
	}

	// test function.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := DeleteFile(testCase.path)
			if testCase.expected == true && err != nil {
				t.Errorf("expected: %v, got: %v ", testCase.expected, err)
			}
			if testCase.expected == false && err == nil {
				t.Errorf("expected error but get nil")
			}

			// check delete
			if testCase.expected {
				_, err := os.Stat(testCase.path)
				if err == nil {
					t.Errorf("check file failed, the DeleteFile() failed to delete file: %v", err)
				}
			}
		})
	}
}

// createFile used to create a new file
func createFile(t *testing.T, name string) (string, string) {
	ex, err := os.Executable()
	if err != nil {
		t.Errorf("failed to get execute file: %v", err)
	}
	path := filepath.Join(filepath.Dir(ex), name)
	file, err := os.Create(path)
	if err != nil {
		t.Errorf("failed to create test file: %v", err)
		return "", ""
	}
	file.Close()
	return ex, path
}

// checkFileContent used to check if the last line is consistent with content.
func checkFileContent(path, content string, t *testing.T) error {
	f, err := os.Open(path)
	if err != nil {
		t.Errorf("failed to open file: %v", err)
		return err
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	lastLine := lines[len(lines)-1]
	if lastLine != content {
		t.Errorf("expected content '%v', got '%v'", content, lastLine)
		return err
	}

	f.Close()

	err = os.Remove(path)
	if err != nil {
		t.Errorf("failed to delete file: %v", err)
		return err
	}

	return nil
}
