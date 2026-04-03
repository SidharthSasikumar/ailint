package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/SidharthSasikumar/ailint/internal/parser"
	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// Scanner walks a directory tree and returns supported source files.
type Scanner struct {
	root    string
	exclude []string
}

// New returns a Scanner for the given root directory.
func New(root string, exclude []string) *Scanner {
	return &Scanner{
		root:    root,
		exclude: exclude,
	}
}

// Scan walks the tree and returns FileContexts for supported files.
func (s *Scanner) Scan() ([]*types.FileContext, error) {
	var files []*types.FileContext

	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		relPath, _ := filepath.Rel(s.root, path)
		if relPath == "." {
			return nil
		}

		// Check exclusions
		if s.isExcluded(relPath, d) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		// Check if we support this file type
		ext := filepath.Ext(path)
		lang := parser.LanguageFromExt(ext)
		if lang == "" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}

		files = append(files, types.NewFileContext(relPath, content, lang))
		return nil
	})

	return files, err
}

// isExcluded returns true if the path matches any exclusion pattern.
func (s *Scanner) isExcluded(relPath string, d fs.DirEntry) bool {
	for _, pattern := range s.exclude {
		// Directory pattern (ends with /)
		if strings.HasSuffix(pattern, "/") {
			dirName := strings.TrimSuffix(pattern, "/")
			if d.IsDir() && d.Name() == dirName {
				return true
			}
			if strings.HasPrefix(relPath, dirName+"/") || relPath == dirName {
				return true
			}
			continue
		}

		// Glob pattern
		if matched, _ := filepath.Match(pattern, d.Name()); matched {
			return true
		}

		// Also match against relative path
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
	}
	return false
}
