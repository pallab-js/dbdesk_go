package sqlite

import "strings"

// SplitStatements splits SQL on top-level semicolons, ignoring semicolons
// inside strings, quoted identifiers, brackets and comments.
func SplitStatements(input string) []string {
	var out []string
	var b strings.Builder
	i, n := 0, len(input)
	flush := func() {
		if s := strings.TrimSpace(b.String()); s != "" {
			out = append(out, s)
		}
		b.Reset()
	}
	for i < n {
		c := input[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			quote := c
			b.WriteByte(c)
			i++
			for i < n {
				b.WriteByte(input[i])
				if input[i] == quote {
					if i+1 < n && input[i+1] == quote {
						b.WriteByte(input[i+1])
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
		case c == '[':
			b.WriteByte(c)
			i++
			for i < n {
				b.WriteByte(input[i])
				if input[i] == ']' {
					i++
					break
				}
				i++
			}
		case c == '-' && i+1 < n && input[i+1] == '-':
			for i < n && input[i] != '\n' {
				b.WriteByte(input[i])
				i++
			}
		case c == '/' && i+1 < n && input[i+1] == '*':
			b.WriteString("/*")
			i += 2
			for i < n {
				if i+1 < n && input[i] == '*' && input[i+1] == '/' {
					b.WriteString("*/")
					i += 2
					break
				}
				b.WriteByte(input[i])
				i++
			}
		case c == ';':
			flush()
			i++
		default:
			b.WriteByte(c)
			i++
		}
	}
	flush()
	return out
}

// FirstKeyword returns the first SQL keyword, skipping whitespace and comments.
func FirstKeyword(statement string) string {
	s := strings.TrimSpace(statement)
	for len(s) > 0 {
		if strings.HasPrefix(s, "--") {
			if i := strings.IndexByte(s, '\n'); i >= 0 {
				s = strings.TrimSpace(s[i+1:])
				continue
			}
			return ""
		}
		if strings.HasPrefix(s, "/*") {
			if i := strings.Index(s, "*/"); i >= 0 {
				s = strings.TrimSpace(s[i+2:])
				continue
			}
			return ""
		}
		break
	}
	end := 0
	for end < len(s) {
		c := s[end]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			end++
			continue
		}
		break
	}
	return strings.ToUpper(s[:end])
}

var readKeywords = map[string]bool{
	"SELECT": true, "WITH": true, "EXPLAIN": true, "PRAGMA": true,
	"VALUES": true, "TABLE": true,
}

func IsReadStatement(keyword string) bool { return readKeywords[keyword] }
