package main

import (
	"path/filepath"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokNormal TokenType = iota
	TokKeyword
	TokString
	TokComment
	TokNumber
	TokType
	TokFunction
	TokOperator
)

type LangDef struct {
	Name           string
	Extensions     []string
	Keywords       map[string]bool
	Types          map[string]bool
	LineComment    string
	BlockCommentOpen  string
	BlockCommentClose string
	StringDelims   []string
	CharDelims     []string
}

var languages []LangDef

func init() {
	languages = []LangDef{
		{
			Name: "go", Extensions: []string{".go"},
			Keywords: map[string]bool{
				"break": true, "case": true, "chan": true, "const": true, "continue": true,
				"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
				"func": true, "go": true, "goto": true, "if": true, "import": true,
				"interface": true, "map": true, "package": true, "range": true, "return": true,
				"select": true, "struct": true, "switch": true, "type": true, "var": true,
			},
			Types: map[string]bool{
				"bool": true, "byte": true, "complex64": true, "complex128": true,
				"error": true, "float32": true, "float64": true, "int": true, "int8": true,
				"int16": true, "int32": true, "int64": true, "rune": true, "string": true,
				"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
				"uintptr": true,
			},
			LineComment: "//", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\"", "`", "`"},
			CharDelims:   []string{"'"},
		},
		{
			Name: "python", Extensions: []string{".py", ".pyw"},
			Keywords: map[string]bool{
				"and": true, "as": true, "assert": true, "async": true, "await": true,
				"break": true, "class": true, "continue": true, "def": true, "del": true,
				"elif": true, "else": true, "except": true, "finally": true, "for": true,
				"from": true, "global": true, "if": true, "import": true, "in": true,
				"is": true, "lambda": true, "nonlocal": true, "not": true, "or": true,
				"pass": true, "raise": true, "return": true, "try": true, "while": true,
				"with": true, "yield": true, "True": true, "False": true, "None": true,
			},
			LineComment: "#", StringDelims: []string{"\"", "'", "\"\"\"", "'''"},
		},
		{
			Name: "javascript", Extensions: []string{".js", ".jsx", ".mjs", ".cjs"},
			Keywords: map[string]bool{
				"async": true, "await": true, "break": true, "case": true, "catch": true,
				"class": true, "const": true, "continue": true, "debugger": true, "default": true,
				"delete": true, "do": true, "else": true, "export": true, "extends": true,
				"finally": true, "for": true, "function": true, "if": true, "import": true,
				"in": true, "instanceof": true, "let": true, "new": true, "of": true,
				"return": true, "super": true, "switch": true, "this": true, "throw": true,
				"try": true, "typeof": true, "var": true, "void": true, "while": true,
				"with": true, "yield": true, "true": true, "false": true, "null": true,
				"undefined": true,
			},
			LineComment: "//", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\"", "'", "`"},
		},
		{
			Name: "typescript", Extensions: []string{".ts", ".tsx", ".mts", ".cts"},
			Keywords: map[string]bool{
				"abstract": true, "as": true, "async": true, "await": true, "break": true,
				"case": true, "catch": true, "class": true, "const": true, "continue": true,
				"debugger": true, "declare": true, "default": true, "delete": true, "do": true,
				"else": true, "enum": true, "export": true, "extends": true, "finally": true,
				"for": true, "function": true, "if": true, "implements": true, "import": true,
				"in": true, "instanceof": true, "interface": true, "keyof": true, "let": true,
				"namespace": true, "new": true, "of": true, "private": true, "protected": true,
				"public": true, "readonly": true, "return": true, "static": true, "super": true,
				"switch": true, "this": true, "throw": true, "try": true, "type": true,
				"typeof": true, "var": true, "void": true, "while": true, "with": true,
				"yield": true, "true": true, "false": true, "null": true, "undefined": true,
			},
			Types: map[string]bool{
				"any": true, "boolean": true, "never": true, "number": true, "object": true,
				"string": true, "symbol": true, "unknown": true, "void": true,
			},
			LineComment: "//", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\"", "'", "`"},
		},
		{
			Name: "c", Extensions: []string{".c", ".h"},
			Keywords: map[string]bool{
				"auto": true, "break": true, "case": true, "char": true, "const": true,
				"continue": true, "default": true, "do": true, "double": true, "else": true,
				"enum": true, "extern": true, "float": true, "for": true, "goto": true,
				"if": true, "inline": true, "int": true, "long": true, "register": true,
				"return": true, "short": true, "signed": true, "sizeof": true, "static": true,
				"struct": true, "switch": true, "typedef": true, "union": true, "unsigned": true,
				"void": true, "volatile": true, "while": true,
			},
			Types: map[string]bool{
				"bool": true, "char": true, "double": true, "float": true, "int": true,
				"long": true, "short": true, "size_t": true, "ssize_t": true, "uint8_t": true,
				"uint16_t": true, "uint32_t": true, "uint64_t": true, "int8_t": true,
				"int16_t": true, "int32_t": true, "int64_t": true, "void": true,
			},
			LineComment: "//", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\""}, CharDelims: []string{"'"},
		},
		{
			Name: "cpp", Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx"},
			Keywords: map[string]bool{
				"alignas": true, "alignof": true, "auto": true, "bool": true, "break": true,
				"case": true, "catch": true, "char": true, "class": true, "const": true,
				"constexpr": true, "continue": true, "decltype": true, "default": true, "delete": true,
				"do": true, "double": true, "else": true, "enum": true, "explicit": true,
				"export": true, "extern": true, "false": true, "float": true, "for": true,
				"friend": true, "goto": true, "if": true, "inline": true, "int": true,
				"long": true, "mutable": true, "namespace": true, "new": true, "noexcept": true,
				"nullptr": true, "operator": true, "override": true, "private": true, "protected": true,
				"public": true, "register": true, "return": true, "short": true, "signed": true,
				"sizeof": true, "static": true, "struct": true, "switch": true, "template": true,
				"this": true, "throw": true, "true": true, "try": true, "typedef": true,
				"typeid": true, "typename": true, "union": true, "unsigned": true, "using": true,
				"virtual": true, "void": true, "volatile": true, "while": true,
			},
			LineComment: "//", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\""}, CharDelims: []string{"'"},
		},
		{
			Name: "rust", Extensions: []string{".rs"},
			Keywords: map[string]bool{
				"as": true, "async": true, "await": true, "break": true, "const": true,
				"continue": true, "crate": true, "dyn": true, "else": true, "enum": true,
				"extern": true, "false": true, "fn": true, "for": true, "if": true,
				"impl": true, "in": true, "let": true, "loop": true, "match": true,
				"mod": true, "move": true, "mut": true, "pub": true, "ref": true,
				"return": true, "self": true, "Self": true, "static": true, "struct": true,
				"super": true, "trait": true, "true": true, "type": true, "unsafe": true,
				"use": true, "where": true, "while": true,
			},
			Types: map[string]bool{
				"bool": true, "char": true, "f32": true, "f64": true, "i8": true,
				"i16": true, "i32": true, "i64": true, "isize": true, "str": true,
				"String": true, "u8": true, "u16": true, "u32": true, "u64": true,
				"usize": true, "Vec": true, "Option": true, "Result": true, "Box": true,
			},
			LineComment: "//", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\""}, CharDelims: []string{"'"},
		},
		{
			Name: "java", Extensions: []string{".java"},
			Keywords: map[string]bool{
				"abstract": true, "assert": true, "boolean": true, "break": true, "byte": true,
				"case": true, "catch": true, "char": true, "class": true, "continue": true,
				"default": true, "do": true, "double": true, "else": true, "enum": true,
				"extends": true, "final": true, "finally": true, "float": true, "for": true,
				"if": true, "implements": true, "import": true, "instanceof": true, "int": true,
				"interface": true, "long": true, "native": true, "new": true, "package": true,
				"private": true, "protected": true, "public": true, "return": true, "short": true,
				"static": true, "strictfp": true, "super": true, "switch": true, "synchronized": true,
				"this": true, "throw": true, "throws": true, "transient": true, "try": true,
				"void": true, "volatile": true, "while": true, "true": true, "false": true,
				"null": true,
			},
			LineComment: "//", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\""}, CharDelims: []string{"'"},
		},
		{
			Name: "ruby", Extensions: []string{".rb"},
			Keywords: map[string]bool{
				"alias": true, "and": true, "begin": true, "break": true, "case": true,
				"class": true, "def": true, "defined": true, "do": true, "else": true,
				"elsif": true, "end": true, "ensure": true, "false": true, "for": true,
				"if": true, "in": true, "module": true, "next": true, "nil": true,
				"not": true, "or": true, "redo": true, "rescue": true, "retry": true,
				"return": true, "self": true, "super": true, "then": true, "true": true,
				"undef": true, "unless": true, "until": true, "when": true, "while": true,
				"yield": true,
			},
			LineComment: "#", StringDelims: []string{"\"", "'"},
		},
		{
			Name: "php", Extensions: []string{".php"},
			Keywords: map[string]bool{
				"abstract": true, "and": true, "array": true, "as": true, "break": true,
				"callable": true, "case": true, "catch": true, "class": true, "clone": true,
				"const": true, "continue": true, "declare": true, "default": true, "die": true,
				"do": true, "echo": true, "else": true, "elseif": true, "empty": true,
				"enddeclare": true, "endfor": true, "endforeach": true, "endif": true, "endswitch": true,
				"endwhile": true, "eval": true, "exit": true, "extends": true, "final": true,
				"finally": true, "fn": true, "for": true, "foreach": true, "function": true,
				"global": true, "goto": true, "if": true, "implements": true, "include": true,
				"include_once": true, "instanceof": true, "insteadof": true, "interface": true, "isset": true,
				"list": true, "match": true, "namespace": true, "new": true, "or": true,
				"print": true, "private": true, "protected": true, "public": true, "readonly": true,
				"require": true, "require_once": true, "return": true, "static": true, "switch": true,
				"throw": true, "trait": true, "try": true, "unset": true, "use": true,
				"var": true, "while": true, "xor": true, "yield": true, "true": true,
				"false": true, "null": true,
			},
			LineComment: "//", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\"", "'"},
		},
		{
			Name: "html", Extensions: []string{".html", ".htm"},
			Keywords:       map[string]bool{},
			LineComment:    "",
			BlockCommentOpen: "<!--", BlockCommentClose: "-->",
			StringDelims: []string{"\""},
		},
		{
			Name: "css", Extensions: []string{".css", ".scss", ".sass", ".less"},
			Keywords:       map[string]bool{},
			LineComment:    "",
			BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\"", "'"},
		},
		{
			Name: "json", Extensions: []string{".json", ".jsonc"},
			Keywords: map[string]bool{
				"true": true, "false": true, "null": true,
			},
			StringDelims: []string{"\""},
		},
		{
			Name: "yaml", Extensions: []string{".yaml", ".yml"},
			Keywords: map[string]bool{
				"true": true, "false": true, "null": true, "yes": true, "no": true,
				"on": true, "off": true,
			},
			LineComment: "#", StringDelims: []string{"\"", "'"},
		},
		{
			Name: "xml", Extensions: []string{".xml", ".svg", ".xaml"},
			BlockCommentOpen: "<!--", BlockCommentClose: "-->",
			StringDelims: []string{"\""},
		},
		{
			Name: "markdown", Extensions: []string{".md", ".mdx", ".markdown"},
		},
		{
			Name: "bash", Extensions: []string{".sh", ".bash", ".zsh", ".fish"},
			Keywords: map[string]bool{
				"if": true, "then": true, "else": true, "elif": true, "fi": true,
				"case": true, "esac": true, "for": true, "while": true, "until": true,
				"do": true, "done": true, "in": true, "function": true, "select": true,
				"time": true, "coproc": true, "declare": true, "local": true, "export": true,
				"readonly": true, "unset": true, "source": true, "exit": true, "return": true,
				"echo": true, "printf": true, "test": true, "eval": true, "exec": true,
			},
			LineComment: "#", StringDelims: []string{"\"", "'"},
		},
		{
			Name: "sql", Extensions: []string{".sql"},
			Keywords: map[string]bool{
				"select": true, "from": true, "where": true, "insert": true, "update": true,
				"delete": true, "create": true, "drop": true, "alter": true, "table": true,
				"into": true, "values": true, "set": true, "join": true, "left": true,
				"right": true, "inner": true, "outer": true, "on": true, "and": true,
				"or": true, "not": true, "null": true, "is": true, "in": true,
				"like": true, "between": true, "order": true, "by": true, "group": true,
				"having": true, "limit": true, "offset": true, "union": true, "all": true,
				"as": true, "exists": true, "distinct": true, "count": true, "sum": true,
				"avg": true, "min": true, "max": true, "primary": true, "key": true,
				"foreign": true, "references": true, "index": true, "unique": true, "check": true,
				"default": true, "cascade": true, "begin": true, "commit": true, "rollback": true,
				"true": true, "false": true,
			},
			LineComment: "--", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"'"},
		},
		{
			Name: "kotlin", Extensions: []string{".kt", ".kts"},
			Keywords: map[string]bool{
				"abstract": true, "annotation": true, "as": true, "break": true, "by": true,
				"catch": true, "class": true, "companion": true, "const": true, "constructor": true,
				"continue": true, "data": true, "do": true, "else": true, "enum": true,
				"false": true, "final": true, "finally": true, "for": true, "fun": true,
				"if": true, "import": true, "in": true, "init": true, "inner": true,
				"interface": true, "internal": true, "is": true, "lateinit": true, "null": true,
				"object": true, "open": true, "operator": true, "out": true, "override": true,
				"package": true, "private": true, "protected": true, "public": true, "return": true,
				"sealed": true, "super": true, "suspend": true, "this": true, "throw": true,
				"true": true, "try": true, "typealias": true, "val": true, "var": true,
				"when": true, "while": true,
			},
			LineComment: "//", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\"", "\"\"\""},
		},
		{
			Name: "swift", Extensions: []string{".swift"},
			Keywords: map[string]bool{
				"as": true, "associatedtype": true, "break": true, "case": true, "catch": true,
				"class": true, "continue": true, "default": true, "defer": true, "deinit": true,
				"do": true, "else": true, "enum": true, "extension": true, "fallthrough": true,
				"false": true, "fileprivate": true, "for": true, "func": true, "guard": true,
				"if": true, "import": true, "in": true, "init": true, "inout": true,
				"internal": true, "is": true, "let": true, "nil": true, "open": true,
				"operator": true, "private": true, "protocol": true, "public": true, "repeat": true,
				"return": true, "self": true, "Self": true, "static": true, "struct": true,
				"subscript": true, "super": true, "switch": true, "throw": true, "throws": true,
				"true": true, "try": true, "typealias": true, "var": true, "where": true,
				"while": true,
			},
			Types: map[string]bool{
				"Bool": true, "Int": true, "Double": true, "Float": true, "String": true,
				"Array": true, "Dictionary": true, "Set": true, "Optional": true, "Void": true,
				"Any": true, "AnyObject": true,
			},
			LineComment: "//", BlockCommentOpen: "/*", BlockCommentClose: "*/",
			StringDelims: []string{"\"", "\"\"\""},
		},
	}
}

func detectLang(filename string) *LangDef {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		base := strings.ToLower(filepath.Base(filename))
		for i := range languages {
			for _, e := range languages[i].Extensions {
				if e == base {
					return &languages[i]
				}
			}
		}
		if base == "makefile" || base == "makefile.in" || base == "gnumakefile" {
			return &LangDef{Name: "makefile", LineComment: "#"}
		}
		if base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") {
			return &LangDef{
				Name: "dockerfile",
				Keywords: map[string]bool{
					"FROM": true, "RUN": true, "CMD": true, "LABEL": true, "EXPOSE": true,
					"ENV": true, "ADD": true, "COPY": true, "ENTRYPOINT": true, "VOLUME": true,
					"USER": true, "WORKDIR": true, "ARG": true, "ONBUILD": true, "STOPSIGNAL": true,
					"HEALTHCHECK": true, "SHELL": true, "MAINTAINER": true,
				},
				LineComment: "#",
			}
		}
		return nil
	}
	for i := range languages {
		for _, e := range languages[i].Extensions {
			if e == ext {
				return &languages[i]
			}
		}
	}
	return nil
}

type Span struct {
	Start, End int
	Type       TokenType
}

func highlight(line string, lang *LangDef, inBlockComment *bool) []Span {
	if lang == nil {
		return nil
	}

	var spans []Span
	runes := []rune(line)
	n := len(runes)
	pos := 0

	if *inBlockComment {
		endIdx := strings.Index(line, lang.BlockCommentClose)
		if endIdx >= 0 {
			endRune := runeCount(line[:endIdx]) + runeCount(lang.BlockCommentClose)
			spans = append(spans, Span{0, endRune, TokComment})
			*inBlockComment = false
			pos = endRune
		} else {
			spans = append(spans, Span{0, n, TokComment})
			return spans
		}
	}

	for pos < n {
		if lang.BlockCommentOpen != "" && matchAt(runes, pos, lang.BlockCommentOpen) {
			endIdx := strings.Index(string(runes[pos+len(lang.BlockCommentOpen):]), lang.BlockCommentClose)
			if endIdx >= 0 {
				endPos := pos + len(lang.BlockCommentOpen) + endIdx + len(lang.BlockCommentClose)
				spans = append(spans, Span{pos, endPos, TokComment})
				pos = endPos
				continue
			} else {
				spans = append(spans, Span{pos, n, TokComment})
				*inBlockComment = true
				return spans
			}
		}

		if lang.LineComment != "" && matchAt(runes, pos, lang.LineComment) {
			spans = append(spans, Span{pos, n, TokComment})
			return spans
		}

		for _, delim := range lang.StringDelims {
			if matchAt(runes, pos, delim) {
				endPos := findStringEnd(runes, pos, delim)
				spans = append(spans, Span{pos, endPos, TokString})
				pos = endPos
				goto next
			}
		}

		for _, delim := range lang.CharDelims {
			if matchAt(runes, pos, delim) {
				endPos := findCharEnd(runes, pos, delim)
				spans = append(spans, Span{pos, endPos, TokString})
				pos = endPos
				goto next
			}
		}

		if unicode.IsDigit(runes[pos]) || (runes[pos] == '.' && pos+1 < n && unicode.IsDigit(runes[pos+1])) {
			start := pos
			for pos < n && (unicode.IsDigit(runes[pos]) || runes[pos] == '.' || runes[pos] == 'x' || runes[pos] == 'X' ||
				(pos > start && (runes[pos] >= 'a' && runes[pos] <= 'f') || (runes[pos] >= 'A' && runes[pos] <= 'F'))) {
				pos++
			}
			spans = append(spans, Span{start, pos, TokNumber})
			continue
		}

		if isIdentStart(runes[pos]) {
			start := pos
			for pos < n && isIdentPart(runes[pos]) {
				pos++
			}
			word := string(runes[start:pos])
			if lang.Keywords != nil && lang.Keywords[word] {
				spans = append(spans, Span{start, pos, TokKeyword})
			} else if lang.Types != nil && lang.Types[word] {
				spans = append(spans, Span{start, pos, TokType})
			} else {
				after := pos
				for after < n && runes[after] == ' ' {
					after++
				}
				if after < n && runes[after] == '(' {
					spans = append(spans, Span{start, pos, TokFunction})
				}
			}
			continue
		}

		pos++
	next:
	}

	return spans
}

func matchAt(runes []rune, pos int, s string) bool {
	if pos+len(s) > len(runes) {
		return false
	}
	for i, ch := range s {
		if runes[pos+i] != ch {
			return false
		}
	}
	return true
}

func findStringEnd(runes []rune, start int, delim string) int {
	pos := start + len(delim)
	escape := false
	for pos < len(runes) {
		if escape {
			escape = false
			pos++
			continue
		}
		if runes[pos] == '\\' {
			escape = true
			pos++
			continue
		}
		if matchAt(runes, pos, delim) {
			return pos + len(delim)
		}
		pos++
	}
	return len(runes)
}

func findCharEnd(runes []rune, start int, delim string) int {
	pos := start + len(delim)
	escape := false
	for pos < len(runes) {
		if escape {
			escape = false
			pos++
			continue
		}
		if runes[pos] == '\\' {
			escape = true
			pos++
			continue
		}
		if matchAt(runes, pos, delim) {
			return pos + len(delim)
		}
		pos++
	}
	if pos > start+len(delim) && pos-start-len(delim) <= 1 {
		return start + len(delim) + 1
	}
	return len(runes)
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isIdentPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}

func runeCount(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}
