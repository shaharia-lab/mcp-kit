package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shaharia-lab/goai/mcp"
)

func TestReadFileTool(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "fs_tools_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testContent := "Hello, World!"
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Prepare test input
	args, err := json.Marshal(map[string]string{
		"path": testFile,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Execute tool
	result, err := fileSystemReadFile.Handler(context.Background(), mcp.CallToolParams{
		Arguments: args,
	})

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(result.Content))
	}
	if result.Content[0].Text != testContent {
		t.Errorf("Expected content %q, got %q", testContent, result.Content[0].Text)
	}
}

func TestWriteFileTool(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_tools_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "write_test.txt")
	testContent := "Test content"

	// Prepare test input
	args, err := json.Marshal(map[string]string{
		"path":    testFile,
		"content": testContent,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Execute tool
	_, err = fileSystemWriteFile.Handler(context.Background(), mcp.CallToolParams{
		Arguments: args,
	})

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify file contents
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != testContent {
		t.Errorf("Expected file content %q, got %q", testContent, string(content))
	}
}

func TestGetFileInfoTool(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_tools_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "info_test.txt")
	testContent := "Test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Prepare test input
	args, err := json.Marshal(map[string]string{
		"path": testFile,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Execute tool
	result, err := fileSystemGetFileInfo.Handler(context.Background(), mcp.CallToolParams{
		Arguments: args,
	})

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var fileInfo FileInfo
	if err := json.Unmarshal([]byte(result.Content[0].Text), &fileInfo); err != nil {
		t.Fatal(err)
	}

	if fileInfo.Size != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), fileInfo.Size)
	}
	if fileInfo.IsDirectory {
		t.Error("Expected IsDirectory to be false")
	}
}

func TestListDirectoryTool(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_tools_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test structure
	testFiles := []string{"file1.txt", "file2.txt"}
	testDirs := []string{"dir1", "dir2"}

	for _, file := range testFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, file), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	for _, dir := range testDirs {
		if err := os.Mkdir(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Prepare test input
	args, err := json.Marshal(map[string]string{
		"path": tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Execute tool
	result, err := fileSystemListDirectory.Handler(context.Background(), mcp.CallToolParams{
		Arguments: args,
	})

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	listing := result.Content[0].Text
	for _, file := range testFiles {
		if !strings.Contains(listing, "[FILE] "+file) {
			t.Errorf("Expected listing to contain %q", "[FILE] "+file)
		}
	}
	for _, dir := range testDirs {
		if !strings.Contains(listing, "[DIR] "+dir) {
			t.Errorf("Expected listing to contain %q", "[DIR] "+dir)
		}
	}
}

// Helper function to create table-driven tests
type toolTest struct {
	name     string
	args     map[string]interface{}
	setup    func(t *testing.T, tmpDir string) string
	validate func(t *testing.T, result mcp.CallToolResult, tmpDir string)
}

func runToolTest(t *testing.T, tool mcp.Tool, tt toolTest) {
	t.Run(tt.name, func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "fs_tools_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// Run setup if provided
		if tt.setup != nil {
			tt.setup(t, tmpDir)
		}

		// Prepare arguments
		args, err := json.Marshal(tt.args)
		if err != nil {
			t.Fatal(err)
		}

		// Execute tool
		result, err := tool.Handler(context.Background(), mcp.CallToolParams{
			Arguments: args,
		})
		if err != nil {
			t.Errorf("Tool execution failed: %v", err)
			return
		}

		// Run validation
		if tt.validate != nil {
			tt.validate(t, result, tmpDir)
		}
	})
}
