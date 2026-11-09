package dbe

import (
	"reflect"
	"testing"
)

func TestSplitStatements(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []string
	}{
		{
			name: "all statements terminated",
			body: "a;\nb;",
			want: []string{"a;", "\nb;"},
		},
		{
			name: "single statement without semicolon",
			body: "-- id: x\n-- options: chop\n\nINSERT INTO t VALUES (1)\n",
			want: []string{"-- id: x\n-- options: chop\n\nINSERT INTO t VALUES (1)\n"},
		},
		{
			name: "final statement without semicolon",
			body: "a;\nb",
			want: []string{"a;", "\nb"},
		},
		{
			name: "trailing whitespace and comments are not a statement",
			body: "a;\n\n-- done; really\n/* the end; */\n",
			want: []string{"a;"},
		},
		{
			name: "comment-only body",
			body: "-- id: x\n/* nothing; here */\n",
			want: nil,
		},
		{
			name: "empty statements dropped",
			body: ";; a;;",
			want: []string{" a;"},
		},
		{
			name: "semicolon in line comment",
			body: "-- first; comment\na;\nb -- trailing; comment\n;",
			want: []string{"-- first; comment\na;", "\nb -- trailing; comment\n;"},
		},
		{
			name: "CR ends line comment",
			body: "-- c\r\na;\r\nb",
			want: []string{"-- c\r\na;", "\r\nb"},
		},
		{
			name: "semicolon in nested block comment",
			body: "/* a; /* b; */ c; */ d; e",
			want: []string{"/* a; /* b; */ c; */ d;", " e"},
		},
		{
			name: "semicolon in string",
			body: "select 'a;b'; c",
			want: []string{"select 'a;b';", " c"},
		},
		{
			name: "doubled quote in string",
			body: "select 'it''s;'; c",
			want: []string{"select 'it''s;';", " c"},
		},
		{
			name: "backslash is literal in plain string",
			body: `select 'a\'; c`,
			want: []string{`select 'a\';`, " c"},
		},
		{
			name: "backslash escapes quote in E string",
			body: `select E'a\';b'; c`,
			want: []string{`select E'a\';b';`, " c"},
		},
		{
			name: "doubled quote then backslash in E string",
			body: `select e'a'';\';b'; c`,
			want: []string{`select e'a'';\';b';`, " c"},
		},
		{
			name: "identifier ending in e is not an E string",
			body: `select type'a\'; c`,
			want: []string{`select type'a\';`, " c"},
		},
		{
			name: "semicolon in quoted identifier",
			body: `select "a;""b"; c`,
			want: []string{`select "a;""b";`, " c"},
		},
		{
			name: "comment markers in string",
			body: "select '--', '/*'; c",
			want: []string{"select '--', '/*';", " c"},
		},
		{
			name: "quotes in comments",
			body: "-- it's\n/* \" */ a; b",
			want: []string{"-- it's\n/* \" */ a;", " b"},
		},
		{
			name: "dollar-quoted function body",
			body: "create function f() as $$ begin; x; end; $$ language plpgsql; b",
			want: []string{"create function f() as $$ begin; x; end; $$ language plpgsql;", " b"},
		},
		{
			name: "tagged dollar quote containing other dollars",
			body: "select $fn$ a; $$ b; $other$ c; $fn$; d",
			want: []string{"select $fn$ a; $$ b; $other$ c; $fn$;", " d"},
		},
		{
			name: "positional parameter is not a dollar quote",
			body: "select $1; select $2",
			want: []string{"select $1;", " select $2"},
		},
		{
			name: "dollar inside identifier is not a dollar quote",
			body: "select foo$bar; select baz$bar",
			want: []string{"select foo$bar;", " select baz$bar"},
		},
		{
			name: "unterminated string reaches the server",
			body: "a; select 'b; c",
			want: []string{"a;", " select 'b; c"},
		},
		{
			name: "unterminated dollar quote reaches the server",
			body: "a; select $$ b; c",
			want: []string{"a;", " select $$ b; c"},
		},
		{
			name: "unterminated block comment reaches the server",
			body: "a; /* b; c",
			want: []string{"a;", " /* b; c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, stmt := range splitStatements([]byte(tt.body)) {
				got = append(got, string(stmt))
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitStatements(%q)\n got: %q\nwant: %q", tt.body, got, tt.want)
			}
		})
	}
}
