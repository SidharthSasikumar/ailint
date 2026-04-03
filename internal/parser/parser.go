package parser

import "github.com/SidharthSasikumar/ailint/pkg/types"

// Parser extracts structured information from source files.
type Parser interface {
	// Language returns the language this parser handles.
	Language() string

	// Extensions returns the file extensions this parser supports.
	Extensions() []string

	// ParseImports extracts import statements from file content.
	ParseImports(content []byte) []types.Import
}

// Registry maps languages to their parsers.
var Registry = map[string]Parser{}

func init() {
	for _, p := range []Parser{
		&GoParser{},
		&PythonParser{},
		&JavaScriptParser{},
	} {
		Registry[p.Language()] = p
	}
}

// ForLanguage returns the parser for the given language, or nil.
func ForLanguage(lang string) Parser {
	return Registry[lang]
}

// LanguageFromExt returns the language for a file extension.
func LanguageFromExt(ext string) string {
	for _, p := range Registry {
		for _, e := range p.Extensions() {
			if e == ext {
				return p.Language()
			}
		}
	}
	return ""
}
