package mcp

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// ToolImplementation represents a tool's code implementation
type ToolImplementation struct {
	Name         string
	MainFile     string
	Dependencies []string
	SourceFiles  []string
	Version      string
}

// CodeHasher handles generating deterministic hashes for tool implementations
type CodeHasher struct {
	hasher hash.Hash
}

// NewCodeHasher creates a new code hasher instance
func NewCodeHasher() *CodeHasher {
	return &CodeHasher{
		hasher: sha256.New(),
	}
}

// GenerateToolHash generates a SHA256 hash representing the complete tool implementation
func (ch *CodeHasher) GenerateToolHash(tool *ToolImplementation) (string, error) {
	ch.hasher.Reset()

	// Hash tool metadata first for consistency
	ch.hashString(fmt.Sprintf("name:%s", tool.Name))
	ch.hashString(fmt.Sprintf("version:%s", tool.Version))

	// Hash main implementation file
	if err := ch.hashFile(tool.MainFile); err != nil {
		return "", fmt.Errorf("failed to hash main file %s: %w", tool.MainFile, err)
	}

	// Hash additional source files in deterministic order
	sortedFiles := make([]string, len(tool.SourceFiles))
	copy(sortedFiles, tool.SourceFiles)
	sort.Strings(sortedFiles)

	for _, file := range sortedFiles {
		if err := ch.hashFile(file); err != nil {
			return "", fmt.Errorf("failed to hash source file %s: %w", file, err)
		}
	}

	// Hash dependencies in deterministic order
	sortedDeps := make([]string, len(tool.Dependencies))
	copy(sortedDeps, tool.Dependencies)
	sort.Strings(sortedDeps)

	for _, dep := range sortedDeps {
		ch.hashString(fmt.Sprintf("dep:%s", dep))
	}

	// Return final hash as hex string
	return fmt.Sprintf("%x", ch.hasher.Sum(nil)), nil
}

// GenerateCodeOnlyHash generates hash based only on source code, ignoring metadata
func (ch *CodeHasher) GenerateCodeOnlyHash(sourceFiles []string) (string, error) {
	ch.hasher.Reset()

	// Sort files for deterministic hashing
	sortedFiles := make([]string, len(sourceFiles))
	copy(sortedFiles, sourceFiles)
	sort.Strings(sortedFiles)

	for _, file := range sortedFiles {
		if err := ch.hashFile(file); err != nil {
			return "", fmt.Errorf("failed to hash file %s: %w", file, err)
		}
	}

	return fmt.Sprintf("%x", ch.hasher.Sum(nil)), nil
}

// GenerateStringHash generates hash from source code provided as string
func (ch *CodeHasher) GenerateStringHash(code string, dependencies []string) string {
	ch.hasher.Reset()

	// Normalize code (remove extra whitespace, consistent line endings)
	normalizedCode := ch.normalizeCode(code)
	ch.hashString(normalizedCode)

	// Hash dependencies in deterministic order
	if len(dependencies) > 0 {
		sortedDeps := make([]string, len(dependencies))
		copy(sortedDeps, dependencies)
		sort.Strings(sortedDeps)

		for _, dep := range sortedDeps {
			ch.hashString(fmt.Sprintf("dep:%s", dep))
		}
	}

	return fmt.Sprintf("%x", ch.hasher.Sum(nil))
}

// hashFile reads a file and adds its content to the hash
func (ch *CodeHasher) hashFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add filename to hash for uniqueness
	ch.hashString(fmt.Sprintf("file:%s", filepath))

	// Read and hash file contents
	if _, err := io.Copy(ch.hasher, file); err != nil {
		return err
	}

	return nil
}

// hashString adds a string to the current hash
func (ch *CodeHasher) hashString(s string) {
	ch.hasher.Write([]byte(s))
}

// normalizeCode normalizes source code for consistent hashing
func (ch *CodeHasher) normalizeCode(code string) string {
	// Convert to Unix line endings
	normalized := strings.ReplaceAll(code, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	// Remove trailing whitespace from each line
	lines := strings.Split(normalized, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	// Remove trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}

// DiscoverSourceFiles automatically discovers source files in a directory
func DiscoverSourceFiles(rootDir string, extensions []string) ([]string, error) {
	var sourceFiles []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if slices.Contains(extensions, ext) {
			sourceFiles = append(sourceFiles, path)
		}

		return nil
	})

	return sourceFiles, err
}

// HashComparison represents the result of comparing two hashes
type HashComparison struct {
	Hash1   string
	Hash2   string
	Match   bool
	Changed bool
}

// CompareHashes compares two tool implementation hashes
func CompareHashes(hash1, hash2 string) HashComparison {
	return HashComparison{
		Hash1:   hash1,
		Hash2:   hash2,
		Match:   hash1 == hash2,
		Changed: hash1 != hash2,
	}
}
