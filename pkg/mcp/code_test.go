package mcp

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper functions

func createTempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "test_*.py")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	tmpFile.Close()
	return tmpFile.Name()
}

func createTempDir(t *testing.T) string {
	tmpDir, err := os.CreateTemp("", "test_project_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir.Name()
}

func cleanup(paths ...string) {
	for _, path := range paths {
		os.RemoveAll(path)
	}
}

// Test NewCodeHasher
func TestNewCodeHasher(t *testing.T) {
	hasher := NewCodeHasher()

	if hasher == nil {
		t.Fatal("NewCodeHasher returned nil")
	}

	if hasher.hasher == nil {
		t.Fatal("CodeHasher.hasher is nil")
	}
}

// Test GenerateStringHash
func TestGenerateStringHash(t *testing.T) {
	hasher := NewCodeHasher()

	tests := []struct {
		name         string
		code         string
		dependencies []string
		expectSame   bool
		compareWith  struct {
			code         string
			dependencies []string
		}
	}{
		{
			name:         "identical code and deps",
			code:         "def hello():\n    print('world')",
			dependencies: []string{"requests==2.28.1"},
			expectSame:   true,
			compareWith: struct {
				code         string
				dependencies []string
			}{
				code:         "def hello():\n    print('world')",
				dependencies: []string{"requests==2.28.1"},
			},
		},
		{
			name:         "same code different deps",
			code:         "def hello():\n    print('world')",
			dependencies: []string{"requests==2.28.1"},
			expectSame:   false,
			compareWith: struct {
				code         string
				dependencies []string
			}{
				code:         "def hello():\n    print('world')",
				dependencies: []string{"requests==2.29.0"},
			},
		},
		{
			name:         "different code same deps",
			code:         "def hello():\n    print('world')",
			dependencies: []string{"requests==2.28.1"},
			expectSame:   false,
			compareWith: struct {
				code         string
				dependencies []string
			}{
				code:         "def goodbye():\n    print('world')",
				dependencies: []string{"requests==2.28.1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := hasher.GenerateStringHash(tt.code, tt.dependencies)
			hash2 := hasher.GenerateStringHash(tt.compareWith.code, tt.compareWith.dependencies)

			if len(hash1) != 64 {
				t.Errorf("Expected hash length 64, got %d", len(hash1))
			}

			if len(hash2) != 64 {
				t.Errorf("Expected hash length 64, got %d", len(hash2))
			}

			same := hash1 == hash2
			if same != tt.expectSame {
				t.Errorf("Expected same=%t, got same=%t\nHash1: %s\nHash2: %s",
					tt.expectSame, same, hash1, hash2)
			}
		})
	}
}

// Test code normalization
func TestCodeNormalization(t *testing.T) {
	hasher := NewCodeHasher()

	// Test that different line endings produce same hash
	code1 := "def hello():\n    print('world')\n"
	code2 := "def hello():\r\n    print('world')\r\n"
	code3 := "def hello():\r    print('world')\r"

	hash1 := hasher.GenerateStringHash(code1, nil)
	hash2 := hasher.GenerateStringHash(code2, nil)
	hash3 := hasher.GenerateStringHash(code3, nil)

	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("Different line endings should produce same hash\nHash1: %s\nHash2: %s\nHash3: %s",
			hash1, hash2, hash3)
	}

	// Test that trailing whitespace is normalized
	codeWithTrailing := "def hello():    \n    print('world')  \n\n\n"
	codeWithoutTrailing := "def hello():\n    print('world')"

	hashTrailing := hasher.GenerateStringHash(codeWithTrailing, nil)
	hashClean := hasher.GenerateStringHash(codeWithoutTrailing, nil)

	if hashTrailing != hashClean {
		t.Errorf("Trailing whitespace should be normalized\nWith trailing: %s\nClean: %s",
			hashTrailing, hashClean)
	}
}

// Test dependency ordering
func TestDependencyOrdering(t *testing.T) {
	hasher := NewCodeHasher()
	code := "def hello(): pass"

	// Same dependencies in different order should produce same hash
	deps1 := []string{"requests==2.28.1", "numpy==1.21.0", "pandas==1.3.0"}
	deps2 := []string{"pandas==1.3.0", "requests==2.28.1", "numpy==1.21.0"}
	deps3 := []string{"numpy==1.21.0", "pandas==1.3.0", "requests==2.28.1"}

	hash1 := hasher.GenerateStringHash(code, deps1)
	hash2 := hasher.GenerateStringHash(code, deps2)
	hash3 := hasher.GenerateStringHash(code, deps3)

	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("Same dependencies in different order should produce same hash\nHash1: %s\nHash2: %s\nHash3: %s",
			hash1, hash2, hash3)
	}
}

// Test hashFile function
func TestHashFile(t *testing.T) {
	hasher := NewCodeHasher()

	content := "def test_function():\n    return 'test'"
	tmpFile := createTempFile(t, content)
	defer cleanup(tmpFile)

	// Reset hasher state
	hasher.hasher.Reset()

	err := hasher.hashFile(tmpFile)
	if err != nil {
		t.Fatalf("hashFile failed: %v", err)
	}

	// Verify that something was hashed
	hash := fmt.Sprintf("%x", hasher.hasher.Sum(nil))
	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}

	// Test with non-existent file
	err = hasher.hashFile("non_existent_file.py")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// Test GenerateCodeOnlyHash
func TestGenerateCodeOnlyHash(t *testing.T) {
	hasher := NewCodeHasher()

	// Create temp files
	content1 := "def func1(): pass"
	content2 := "def func2(): pass"

	file1 := createTempFile(t, content1)
	file2 := createTempFile(t, content2)
	defer cleanup(file1, file2)

	files := []string{file1, file2}

	hash, err := hasher.GenerateCodeOnlyHash(files)
	if err != nil {
		t.Fatalf("GenerateCodeOnlyHash failed: %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}

	// Test that file order doesn't matter
	filesReversed := []string{file2, file1}
	hashReversed, err := hasher.GenerateCodeOnlyHash(filesReversed)
	if err != nil {
		t.Fatalf("GenerateCodeOnlyHash failed: %v", err)
	}

	if hash != hashReversed {
		t.Errorf("File order should not affect hash\nOriginal: %s\nReversed: %s", hash, hashReversed)
	}
}

// Test GenerateToolHash
func TestGenerateToolHash(t *testing.T) {
	hasher := NewCodeHasher()

	// Create temp files
	mainContent := "def main_function(): pass"
	helperContent := "def helper_function(): pass"

	mainFile := createTempFile(t, mainContent)
	helperFile := createTempFile(t, helperContent)
	defer cleanup(mainFile, helperFile)

	tool := &ToolImplementation{
		Name:         "test_tool",
		MainFile:     mainFile,
		SourceFiles:  []string{helperFile},
		Dependencies: []string{"requests==2.28.1"},
		Version:      "1.0.0",
	}

	hash, err := hasher.GenerateToolHash(tool)
	if err != nil {
		t.Fatalf("GenerateToolHash failed: %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}

	// Test that changing version changes hash
	tool2 := *tool
	tool2.Version = "1.0.1"

	hash2, err := hasher.GenerateToolHash(&tool2)
	if err != nil {
		t.Fatalf("GenerateToolHash failed for tool2: %v", err)
	}

	if hash == hash2 {
		t.Error("Different versions should produce different hashes")
	}

	// Test with non-existent main file
	toolBad := &ToolImplementation{
		Name:     "bad_tool",
		MainFile: "non_existent.py",
		Version:  "1.0.0",
	}

	_, err = hasher.GenerateToolHash(toolBad)
	if err == nil {
		t.Error("Expected error for non-existent main file")
	}
}

// Test DiscoverSourceFiles
func TestDiscoverSourceFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	defer cleanup(tmpDir)

	// Create test files
	files := map[string]string{
		"main.py":       "def main(): pass",
		"helper.py":     "def helper(): pass",
		"config.js":     "const config = {};",
		"readme.txt":    "This is readme",
		"test.go":       "package main",
		"subdir/sub.py": "def sub(): pass",
	}

	for filename, content := range files {
		fullPath := filepath.Join(tmpDir, filename)

		// Create directory if needed
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := ioutil.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Test discovering Python files
	pythonFiles, err := DiscoverSourceFiles(tmpDir, []string{".py"})
	if err != nil {
		t.Fatalf("DiscoverSourceFiles failed: %v", err)
	}

	expectedPython := 3 // main.py, helper.py, subdir/sub.py
	if len(pythonFiles) != expectedPython {
		t.Errorf("Expected %d Python files, got %d: %v", expectedPython, len(pythonFiles), pythonFiles)
	}

	// Test discovering multiple extensions
	multiFiles, err := DiscoverSourceFiles(tmpDir, []string{".py", ".js", ".go"})
	if err != nil {
		t.Fatalf("DiscoverSourceFiles failed: %v", err)
	}

	expectedMulti := 5 // All except readme.txt
	if len(multiFiles) != expectedMulti {
		t.Errorf("Expected %d files, got %d: %v", expectedMulti, len(multiFiles), multiFiles)
	}

	// Verify no .txt files were included
	for _, file := range multiFiles {
		if strings.HasSuffix(file, ".txt") {
			t.Errorf("Should not include .txt files, but found: %s", file)
		}
	}
}

// Test CompareHashes
func TestCompareHashes(t *testing.T) {
	hash1 := "abc123"
	hash2 := "def456"
	hash3 := "abc123"

	// Test different hashes
	comparison1 := CompareHashes(hash1, hash2)
	if comparison1.Match {
		t.Error("Different hashes should not match")
	}
	if !comparison1.Changed {
		t.Error("Different hashes should be marked as changed")
	}
	if comparison1.Hash1 != hash1 || comparison1.Hash2 != hash2 {
		t.Error("Hash values not properly stored in comparison")
	}

	// Test same hashes
	comparison2 := CompareHashes(hash1, hash3)
	if !comparison2.Match {
		t.Error("Same hashes should match")
	}
	if comparison2.Changed {
		t.Error("Same hashes should not be marked as changed")
	}
}

// Test normalizeCode function
func TestNormalizeCode(t *testing.T) {
	hasher := NewCodeHasher()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unix line endings",
			input:    "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "windows line endings",
			input:    "line1\r\nline2\r\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "mac line endings",
			input:    "line1\rline2\rline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "trailing whitespace",
			input:    "line1   \nline2\t\nline3  ",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "trailing empty lines",
			input:    "line1\nline2\n\n\n",
			expected: "line1\nline2",
		},
		{
			name:     "mixed issues",
			input:    "line1  \r\nline2\t\r\nline3   \n\n\n",
			expected: "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasher.normalizeCode(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeCode failed\nInput: %q\nExpected: %q\nGot: %q",
					tt.input, tt.expected, result)
			}
		})
	}
}

// Test hashString function
func TestHashString(t *testing.T) {
	hasher := NewCodeHasher()

	// Reset and hash a string
	hasher.hasher.Reset()
	hasher.hashString("test string")

	hash1 := fmt.Sprintf("%x", hasher.hasher.Sum(nil))

	// Reset and hash the same string
	hasher.hasher.Reset()
	hasher.hashString("test string")

	hash2 := fmt.Sprintf("%x", hasher.hasher.Sum(nil))

	if hash1 != hash2 {
		t.Error("Same string should produce same hash")
	}

	// Hash different string
	hasher.hasher.Reset()
	hasher.hashString("different string")

	hash3 := fmt.Sprintf("%x", hasher.hasher.Sum(nil))

	if hash1 == hash3 {
		t.Error("Different strings should produce different hashes")
	}
}

// Benchmark tests
func BenchmarkGenerateStringHash(b *testing.B) {
	hasher := NewCodeHasher()
	code := `
def web_search(query: str) -> dict:
    """Search the web using an API"""
    import requests
    
    response = requests.get(f"https://api.search.com/search?q={query}")
    return response.json()

def validate_query(query: str) -> bool:
    """Validate search query"""
    return len(query) > 0 and len(query) < 200
`
	dependencies := []string{"requests==2.28.1", "urllib3==1.26.12"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.GenerateStringHash(code, dependencies)
	}
}

func BenchmarkNormalizeCode(b *testing.B) {
	hasher := NewCodeHasher()
	code := "line1  \r\nline2\t\r\nline3   \n\n\n" + strings.Repeat("def func(): pass\r\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.normalizeCode(code)
	}
}

// Integration test
func TestIntegration(t *testing.T) {
	hasher := NewCodeHasher()

	// Create a complete tool
	tmpDir := createTempDir(t)
	defer cleanup(tmpDir)

	mainContent := `
def main_tool(input_data):
    """Main tool function"""
    result = process_data(input_data)
    return result
`

	helperContent := `
def process_data(data):
    """Helper function"""
    return data.upper()
`

	mainPath := filepath.Join(tmpDir, "main.py")
	helperPath := filepath.Join(tmpDir, "helper.py")

	if err := ioutil.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	if err := ioutil.WriteFile(helperPath, []byte(helperContent), 0644); err != nil {
		t.Fatalf("Failed to write helper file: %v", err)
	}

	tool := &ToolImplementation{
		Name:         "integration_test_tool",
		MainFile:     mainPath,
		SourceFiles:  []string{helperPath},
		Dependencies: []string{"requests==2.28.1"},
		Version:      "1.0.0",
	}

	// Generate hash
	hash1, err := hasher.GenerateToolHash(tool)
	if err != nil {
		t.Fatalf("Failed to generate tool hash: %v", err)
	}

	// Generate again - should be identical
	hash2, err := hasher.GenerateToolHash(tool)
	if err != nil {
		t.Fatalf("Failed to generate tool hash second time: %v", err)
	}

	if hash1 != hash2 {
		t.Error("Same tool should produce same hash consistently")
	}

	// Modify helper file slightly
	modifiedHelperContent := strings.ReplaceAll(helperContent, "upper()", "lower()")
	if err := ioutil.WriteFile(helperPath, []byte(modifiedHelperContent), 0644); err != nil {
		t.Fatalf("Failed to write modified helper file: %v", err)
	}

	hash3, err := hasher.GenerateToolHash(tool)
	if err != nil {
		t.Fatalf("Failed to generate tool hash after modification: %v", err)
	}

	if hash1 == hash3 {
		t.Error("Modified tool should produce different hash")
	}

	// Test comparison
	comparison := CompareHashes(hash1, hash3)
	if !comparison.Changed {
		t.Error("Comparison should detect change")
	}
}
