package parser

import "github.com/SidharthSasikumar/ailint/pkg/types"

// Parser extracts imports and structure from source files.
type Parser interface {
	Language() string
	Extensions() []string
	ParseImports(content []byte) []types.Import
}

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

// ForLanguage returns the parser for lang, or nil.
func ForLanguage(lang string) Parser {
	return Registry[lang]
}

// LanguageFromExt maps a file extension to a language name.
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
