package dbe

// splitStatements chops a Postgres script into individual statements for the
// "chop" option. A ';' only ends a statement when it is outside comments,
// quoted strings, quoted identifiers and dollar-quoted strings. The last
// statement doesn't need a ';'.
//
// Each returned statement keeps its original bytes, including any leading
// comments and its ';'. Pieces with no SQL in them (only whitespace and
// comments) are dropped. Unterminated strings and comments count as SQL, so
// they reach the server and fail there instead of being silently skipped.
//
// BEGIN ATOMIC function bodies are not recognized: their inner ';' still split.
func splitStatements(body []byte) [][]byte {
	var stmts [][]byte
	start := 0
	hasSQL := false

	endStatement := func(end int) {
		if hasSQL {
			stmts = append(stmts, body[start:end])
		}
		start = end
		hasSQL = false
	}

	for i := 0; i < len(body); {
		c := body[i]
		switch {
		case c == ';':
			i++
			endStatement(i)
		case isSQLSpace(c):
			i++
		case hasPrefixAt(body, i, "--"):
			i = skipLineComment(body, i)
		case hasPrefixAt(body, i, "/*"):
			end, terminated := skipBlockComment(body, i)
			i = end
			if !terminated {
				hasSQL = true
			}
		case c == '\'':
			hasSQL = true
			i = skipQuoted(body, i, isEscapeStringStart(body, i))
		case c == '"':
			hasSQL = true
			i = skipQuoted(body, i, false)
		case c == '$':
			hasSQL = true
			i = skipDollar(body, i)
		default:
			hasSQL = true
			i++
		}
	}
	endStatement(len(body))

	return stmts
}

func hasPrefixAt(body []byte, i int, prefix string) bool {
	return len(body)-i >= len(prefix) && string(body[i:i+len(prefix)]) == prefix
}

func isSQLSpace(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	}
	return false
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c >= 0x80
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || c == '$' || (c >= '0' && c <= '9')
}

// skipLineComment returns the index of the newline ending the "--" comment at i.
func skipLineComment(body []byte, i int) int {
	for i < len(body) && body[i] != '\n' && body[i] != '\r' {
		i++
	}
	return i
}

// skipBlockComment returns the index just past the "/*" comment at i, which
// may nest, and whether the comment was closed before the end of body.
func skipBlockComment(body []byte, i int) (int, bool) {
	depth := 0
	for i < len(body) {
		switch {
		case hasPrefixAt(body, i, "/*"):
			depth++
			i += 2
		case hasPrefixAt(body, i, "*/"):
			depth--
			i += 2
			if depth == 0 {
				return i, true
			}
		default:
			i++
		}
	}
	return i, false
}

// isEscapeStringStart reports whether the quote at i opens an E'...' string,
// the only kind of string in which a backslash can escape a quote.
func isEscapeStringStart(body []byte, i int) bool {
	if i == 0 || (body[i-1] != 'e' && body[i-1] != 'E') {
		return false
	}
	return i == 1 || !isIdentPart(body[i-2])
}

// skipQuoted returns the index just past the quoted text at i. A doubled quote
// character is an escaped quote, not the end of the text.
func skipQuoted(body []byte, i int, backslashEscapes bool) int {
	quote := body[i]
	for j := i + 1; j < len(body); j++ {
		switch {
		case backslashEscapes && body[j] == '\\':
			j++
		case body[j] == quote:
			if j+1 < len(body) && body[j+1] == quote {
				j++
				continue
			}
			return j + 1
		}
	}
	return len(body)
}

// skipDollar returns the index just past the dollar-quoted string at i. If the
// '$' doesn't open one (as in $1 or foo$bar) it returns i+1.
func skipDollar(body []byte, i int) int {
	if i > 0 && isIdentPart(body[i-1]) {
		return i + 1
	}

	tagEnd := i + 1
	if tagEnd < len(body) && isIdentStart(body[tagEnd]) {
		for tagEnd < len(body) && isIdentPart(body[tagEnd]) && body[tagEnd] != '$' {
			tagEnd++
		}
	}
	if tagEnd >= len(body) || body[tagEnd] != '$' {
		return i + 1
	}

	delimiter := string(body[i : tagEnd+1])
	for j := tagEnd + 1; j < len(body); j++ {
		if hasPrefixAt(body, j, delimiter) {
			return j + len(delimiter)
		}
	}
	return len(body)
}
