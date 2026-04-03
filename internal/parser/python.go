package parser

import (
	"regexp"
	"strings"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// PythonParser handles Python source files.
type PythonParser struct{}

func (p *PythonParser) Language() string     { return "python" }
func (p *PythonParser) Extensions() []string { return []string{".py"} }

var pyImportDirect = regexp.MustCompile(`^import\s+(.+)$`)
var pyFromImport = regexp.MustCompile(`^from\s+(\S+)\s+import`)

func (p *PythonParser) ParseImports(content []byte) []types.Import {
	var imports []types.Import
	for i, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		if m := pyFromImport.FindStringSubmatch(trimmed); m != nil {
			// from X import Y — extract top-level module
			topLevel := strings.Split(m[1], ".")[0]
			if strings.HasPrefix(m[1], ".") {
				continue // Relative imports are fine
			}
			imports = append(imports, types.Import{
				Path: m[1],
				Name: topLevel,
				Line: i + 1,
			})
		} else if m := pyImportDirect.FindStringSubmatch(trimmed); m != nil {
			// import X, Y, Z
			for _, name := range strings.Split(m[1], ",") {
				name = strings.TrimSpace(name)
				// Handle "import X as Y"
				if idx := strings.Index(name, " as "); idx > 0 {
					name = name[:idx]
				}
				if name == "" {
					continue
				}
				topLevel := strings.Split(name, ".")[0]
				imports = append(imports, types.Import{
					Path: name,
					Name: topLevel,
					Line: i + 1,
				})
			}
		}
	}
	return imports
}

// PythonStdlib lists Python 3.x standard library modules.
var PythonStdlib = map[string]bool{
	"__future__": true, "__main__": true, "_thread": true,
	"abc": true, "aifc": true, "argparse": true, "array": true, "ast": true,
	"asynchat": true, "asyncio": true, "asyncore": true, "atexit": true, "audioop": true,
	"base64": true, "bdb": true, "binascii": true, "binhex": true, "bisect": true,
	"builtins": true, "bz2": true,
	"calendar": true, "cgi": true, "cgitb": true, "chunk": true, "cmath": true,
	"cmd": true, "code": true, "codecs": true, "codeop": true, "collections": true,
	"colorsys": true, "compileall": true, "concurrent": true, "configparser": true,
	"contextlib": true, "contextvars": true, "copy": true, "copyreg": true,
	"cProfile": true, "crypt": true, "csv": true, "ctypes": true, "curses": true,
	"dataclasses": true, "datetime": true, "dbm": true, "decimal": true,
	"difflib": true, "dis": true, "distutils": true, "doctest": true,
	"email": true, "encodings": true, "enum": true, "errno": true,
	"faulthandler": true, "fcntl": true, "filecmp": true, "fileinput": true,
	"fnmatch": true, "fractions": true, "ftplib": true, "functools": true,
	"gc": true, "getopt": true, "getpass": true, "gettext": true, "glob": true,
	"graphlib": true, "grp": true, "gzip": true,
	"hashlib": true, "heapq": true, "hmac": true, "html": true, "http": true,
	"idlelib": true, "imaplib": true, "imghdr": true, "imp": true,
	"importlib": true, "inspect": true, "io": true, "ipaddress": true,
	"itertools": true,
	"json":      true,
	"keyword":   true,
	"lib2to3":   true, "linecache": true, "locale": true, "logging": true, "lzma": true,
	"mailbox": true, "mailcap": true, "marshal": true, "math": true,
	"mimetypes": true, "mmap": true, "modulefinder": true, "multiprocessing": true,
	"netrc": true, "nis": true, "nntplib": true, "numbers": true,
	"operator": true, "optparse": true, "os": true, "ossaudiodev": true,
	"pathlib": true, "pdb": true, "pickle": true, "pickletools": true,
	"pipes": true, "pkgutil": true, "platform": true, "plistlib": true,
	"poplib": true, "posix": true, "posixpath": true, "pprint": true,
	"profile": true, "pstats": true, "pty": true, "pwd": true,
	"py_compile": true, "pyclbr": true, "pydoc": true,
	"queue": true, "quopri": true,
	"random": true, "re": true, "readline": true, "reprlib": true,
	"resource": true, "rlcompleter": true, "runpy": true,
	"sched": true, "secrets": true, "select": true, "selectors": true,
	"shelve": true, "shlex": true, "shutil": true, "signal": true,
	"site": true, "smtpd": true, "smtplib": true, "sndhdr": true,
	"socket": true, "socketserver": true, "sqlite3": true, "ssl": true,
	"stat": true, "statistics": true, "string": true, "stringprep": true,
	"struct": true, "subprocess": true, "sunau": true, "symtable": true,
	"sys": true, "sysconfig": true, "syslog": true,
	"tabnanny": true, "tarfile": true, "telnetlib": true, "tempfile": true,
	"termios": true, "test": true, "textwrap": true, "threading": true,
	"time": true, "timeit": true, "tkinter": true, "token": true,
	"tokenize": true, "tomllib": true, "trace": true, "traceback": true,
	"tracemalloc": true, "tty": true, "turtle": true, "turtledemo": true,
	"types": true, "typing": true,
	"unicodedata": true, "unittest": true, "urllib": true, "uu": true, "uuid": true,
	"venv":     true,
	"warnings": true, "wave": true, "weakref": true, "webbrowser": true,
	"winreg": true, "winsound": true, "wsgiref": true,
	"xdrlib": true, "xml": true, "xmlrpc": true,
	"zipapp": true, "zipfile": true, "zipimport": true, "zlib": true, "zoneinfo": true,
}
