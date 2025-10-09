package importer

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestNormalizeQualifiedIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "   ", ""},
		{"simple", "public.widgets", "public.widgets"},
		{"mixed case", "Public.Widgets", "public.widgets"},
		{"quoted", `"Public"."Widgets"`, "public.widgets"},
		{"surrounding whitespace", "  public . Widgets  ", "public.widgets"},
		{"single identifier", "Widgets", "widgets"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeQualifiedIdentifier(tc.in)
			if got != tc.want {
				t.Fatalf("normalizeQualifiedIdentifier(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestShouldExcludeIdentifier(t *testing.T) {
	t.Parallel()

	skip := map[string]struct{}{
		"api_keys":           {},
		"api_key_signatures": {},
	}

	if !shouldExcludeIdentifier("public.api_keys", skip) {
		t.Fatal("expected fully qualified identifier to be excluded")
	}

	if !shouldExcludeIdentifier("api_key_signatures", skip) {
		t.Fatal("expected bare identifier to be excluded")
	}

	if !shouldExcludeIdentifier("public.api_keys_id_seq", skip) {
		t.Fatal("expected related sequence to be excluded")
	}

	if shouldExcludeIdentifier("public.widgets", skip) {
		t.Fatal("did not expect unrelated identifier to be excluded")
	}
}

func TestShouldSkipStatement(t *testing.T) {
	t.Parallel()

	skip := map[string]struct{}{"api_keys": {}, "api_key_signatures": {}}
	stmt := statement{sql: "INSERT INTO public.api_keys VALUES (1);"}

	skipped, reason := shouldSkipStatement(stmt, skip)
	if !skipped {
		t.Fatal("expected statement to be skipped")
	}
	if !strings.Contains(reason, "api_keys") {
		t.Fatalf("expected reason to mention api_keys, got %q", reason)
	}

	stmt = statement{sql: "INSERT INTO public.widgets VALUES (1);"}
	if skipped, _ := shouldSkipStatement(stmt, skip); skipped {
		t.Fatal("did not expect widgets insert to be skipped")
	}
}

func TestShouldSkipExecError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		sql       string
		err       error
		wantSkip  bool
		wantMatch string
	}{
		{
			name:      "duplicate schema_migrations row",
			sql:       "INSERT INTO schema_migrations VALUES (3, false);",
			err:       &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint \"schema_migrations_pkey\""},
			wantSkip:  true,
			wantMatch: "duplicate migration row",
		},
		{
			name:      "primary key already exists",
			sql:       "ALTER TABLE public.widgets ADD CONSTRAINT widgets_pkey PRIMARY KEY (id);",
			err:       &pgconn.PgError{Code: "42P16", Message: "multiple primary keys for table are not allowed"},
			wantSkip:  true,
			wantMatch: "primary key",
		},
		{
			name:      "owner change missing role",
			sql:       "ALTER TABLE public.widgets OWNER TO missing_role;",
			err:       &pgconn.PgError{Code: "42704", Message: "role \"missing_role\" does not exist"},
			wantSkip:  true,
			wantMatch: "owner skipped",
		},
		{
			name:      "object already exists",
			sql:       "CREATE TABLE public.widgets (id INT);",
			err:       &pgconn.PgError{Code: "42P07", Message: "relation \"widgets\" already exists"},
			wantSkip:  true,
			wantMatch: "object already exists",
		},
		{
			name:      "missing relation",
			sql:       "DROP TABLE public.widgets;",
			err:       &pgconn.PgError{Code: "42P01", Message: "relation \"widgets\" does not exist"},
			wantSkip:  true,
			wantMatch: "relation skipped",
		},
		{
			name:     "non skip error",
			sql:      "INSERT INTO public.widgets VALUES (1);",
			err:      errors.New("some other error"),
			wantSkip: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			skipped, reason := shouldSkipExecError(statement{sql: tc.sql}, tc.err)
			if skipped != tc.wantSkip {
				t.Fatalf("skip = %v, want %v", skipped, tc.wantSkip)
			}
			if tc.wantSkip && !strings.Contains(reason, tc.wantMatch) {
				t.Fatalf("reason %q does not include %q", reason, tc.wantMatch)
			}
			if !tc.wantSkip && reason != "" {
				t.Fatalf("expected empty reason, got %q", reason)
			}
		})
	}
}

func TestStatementTarget(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		sql       string
		wantTable string
		wantOp    string
	}{
		{
			name:      "insert",
			sql:       "INSERT INTO public.widgets (id) VALUES (1);",
			wantTable: "public.widgets",
			wantOp:    "INSERT",
		},
		{
			name:      "copy",
			sql:       "COPY ONLY public.widgets FROM STDIN;",
			wantTable: "public.widgets",
			wantOp:    "COPY",
		},
		{
			name:      "alter table",
			sql:       "ALTER TABLE \"Public\".\"Widgets\" ADD COLUMN name TEXT;",
			wantTable: `"Public"."Widgets"`,
			wantOp:    "ALTER TABLE",
		},
		{
			name:      "with cte",
			sql:       "WITH existing AS (SELECT 1) INSERT INTO foo(id) VALUES (1);",
			wantTable: "foo",
			wantOp:    "INSERT",
		},
		{
			name:      "drop index",
			sql:       "DROP INDEX IF EXISTS public.widgets_name_key;",
			wantTable: "public.widgets_name_key",
			wantOp:    "DROP INDEX",
		},
		{
			name:      "update",
			sql:       "UPDATE public.widgets SET name = 'new' WHERE id = 1;",
			wantTable: "public.widgets",
			wantOp:    "UPDATE",
		},
		{
			name:      "delete",
			sql:       "DELETE FROM ONLY public.widgets WHERE id = 1;",
			wantTable: "public.widgets",
			wantOp:    "DELETE",
		},
		{
			name:      "alter index",
			sql:       "ALTER INDEX public.widgets_idx RENAME TO widgets_idx_new;",
			wantTable: "public.widgets_idx",
			wantOp:    "ALTER INDEX",
		},
		{
			name:      "drop table",
			sql:       "DROP TABLE IF EXISTS public.widgets;",
			wantTable: "public.widgets",
			wantOp:    "DROP TABLE",
		},
		{
			name:      "drop sequence",
			sql:       "DROP SEQUENCE IF EXISTS public.widgets_id_seq;",
			wantTable: "public.widgets_id_seq",
			wantOp:    "DROP SEQUENCE",
		},
		{
			name:      "select setval",
			sql:       "SELECT pg_catalog.setval('public.widgets_id_seq', 42, true);",
			wantTable: "public.widgets_id_seq",
			wantOp:    "SELECT",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotTable, gotOp := statementTarget(tc.sql)
			if gotTable != tc.wantTable {
				t.Fatalf("statementTarget table = %q, want %q", gotTable, tc.wantTable)
			}
			if gotOp != tc.wantOp {
				t.Fatalf("statementTarget op = %q, want %q", gotOp, tc.wantOp)
			}
		})
	}
}

func TestFirstIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		input     string
		allowOnly bool
		want      string
	}{
		{"simple", "widgets", false, "widgets"},
		{"only keyword", "ONLY public.widgets", true, "public.widgets"},
		{"if not exists", "IF NOT EXISTS public.widgets", false, "public.widgets"},
		{"quoted", `"Public"."Widgets" remaining`, false, `"Public"."Widgets"`},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := firstIdentifier(tc.input, tc.allowOnly); got != tc.want {
				t.Fatalf("firstIdentifier(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestTrimLeadingKeyword(t *testing.T) {
	t.Parallel()

	if next, ok := trimLeadingKeyword("  IF NOT EXISTS widgets", "IF NOT EXISTS"); !ok || strings.TrimSpace(next) != "widgets" {
		t.Fatalf("expected to trim leading keyword, got ok=%v next=%q", ok, next)
	}

	if _, ok := trimLeadingKeyword("DIFFERENT widgets", "IF NOT EXISTS"); ok {
		t.Fatal("did not expect to trim unrelated prefix")
	}
}

func TestSkipLeadingCTESections(t *testing.T) {
	t.Parallel()

	sql := "WITH cte AS (SELECT 1) INSERT INTO public.widgets VALUES (1);"
	idx := skipLeadingCTESections([]byte(sql), 0)
	if !strings.HasPrefix(sql[idx:], "INSERT") {
		t.Fatalf("expected cursor to point at INSERT, got %q", sql[idx:])
	}

	sql = "SELECT * FROM widgets;"
	if idx := skipLeadingCTESections([]byte(sql), 0); idx != 0 {
		t.Fatalf("expected index to remain unchanged, got %d", idx)
	}
}

func TestFormatSnippet(t *testing.T) {
	t.Parallel()

	if got := formatSnippet([]byte("  \n\t")); got != "<empty>" {
		t.Fatalf("expected empty snippet, got %q", got)
	}

	long := strings.Repeat("a", 200)
	if got := formatSnippet([]byte(long)); !strings.HasSuffix(got, "…") || len([]rune(got)) != 120 {
		t.Fatalf("expected snippet to be truncated to 120 runes with ellipsis, got %q", got)
	}
}

func TestLineAndColumn(t *testing.T) {
	t.Parallel()

	src := []byte("first\nsecond\r\nthird")
	line, column := lineAndColumn(src, strings.Index(string(src), "third"))
	if line != 3 || column != 1 {
		t.Fatalf("expected third line column 1, got line=%d column=%d", line, column)
	}

	line, column = lineAndColumn(src, len(src)+10)
	if line != 3 || column != len("third")+1 {
		t.Fatalf("expected clamped index, got line=%d column=%d", line, column)
	}
}

func TestReadDollarTag(t *testing.T) {
	t.Parallel()

	if tag, ok := readDollarTag([]byte("$tag$")); !ok || tag != "tag" {
		t.Fatalf("expected to read tag, got ok=%v tag=%q", ok, tag)
	}

	if _, ok := readDollarTag([]byte("$invalid-")); ok {
		t.Fatal("did not expect to parse invalid tag")
	}
}

func TestHasDollarTag(t *testing.T) {
	t.Parallel()

	if !hasDollarTag([]byte("$test$rest"), "test") {
		t.Fatal("expected prefix tag to be detected")
	}

	if hasDollarTag([]byte("$test$rest"), "other") {
		t.Fatal("did not expect mismatched tag to be detected")
	}
}

func TestIsCopyFromStdin(t *testing.T) {
	t.Parallel()

	if !isCopyFromStdin("COPY public.widgets FROM STDIN;") {
		t.Fatal("expected COPY FROM STDIN to be detected")
	}

	if isCopyFromStdin("INSERT INTO widgets VALUES (1);") {
		t.Fatal("did not expect non COPY statement to be detected")
	}
}

func TestExtractCopyData(t *testing.T) {
	t.Parallel()

	data := []byte("1\n2\n\\.\nTRAILING")
	copyEnd, advance, err := extractCopyData(data)
	if err != nil {
		t.Fatalf("extractCopyData unexpected error: %v", err)
	}
	if string(data[:copyEnd]) != "1\n2\n" {
		t.Fatalf("expected copy payload to exclude terminator, got %q", data[:copyEnd])
	}
	if advance != len("1\n2\n\\.\n") {
		t.Fatalf("expected advance to include terminator line, got %d", advance)
	}

	if _, _, err := extractCopyData([]byte("no terminator")); err == nil {
		t.Fatal("expected error when terminator is missing")
	}
}

func TestParseStatementsHandlesCopy(t *testing.T) {
	t.Parallel()

	sql := strings.Join([]string{
		"CREATE TABLE widgets (id INT);",
		"COPY widgets (id) FROM stdin;",
		"1",
		"2",
		"\\.",
		"",
	}, "\n")

	stmts, err := parseStatements([]byte(sql))
	if err != nil {
		t.Fatalf("parseStatements unexpected error: %v", err)
	}

	if len(stmts) != 2 {
		t.Fatalf("expected two statements, got %d", len(stmts))
	}

	if stmts[1].sql != "COPY widgets (id) FROM stdin" || !stmts[1].isCopy {
		t.Fatalf("expected COPY statement with copy flag, got %+v", stmts[1])
	}

	if string(stmts[1].copyData) != "1\n2\n" {
		t.Fatalf("unexpected copy data %q", string(stmts[1].copyData))
	}
}

func TestValidateFilePathWithinStorageDir(t *testing.T) {
	t.Parallel()

	cleaned, err := validateFilePath("example.sql")
	if err != nil {
		t.Fatalf("unexpected error validating path: %v", err)
	}

	wantPrefix := storageSQLDir + string(os.PathSeparator)
	if !strings.HasPrefix(cleaned, wantPrefix) || !strings.HasSuffix(cleaned, "example.sql") {
		t.Fatalf("expected path to reside within storage directory, got %q", cleaned)
	}
}

func TestValidateFilePathRejectsTraversal(t *testing.T) {
	t.Parallel()

	if _, err := validateFilePath("../outside.sql"); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
}

func TestExtractSetvalIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{"SELECT pg_catalog.setval('public.widgets_id_seq', 42, true);", "public.widgets_id_seq"},
		{"SELECT setval('widgets_id_seq', 5);", "widgets_id_seq"},
		{"SELECT 1;", ""},
	}

	for _, tc := range cases {
		if got := extractSetvalIdentifier(tc.input); got != tc.want {
			t.Fatalf("extractSetvalIdentifier(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSkipIgnorableSections(t *testing.T) {
	t.Parallel()

	sql := "-- comment\n/* block */\\connect foo\n  SELECT"
	idx := skipIgnorableSections([]byte(sql), 0)
	if !strings.HasPrefix(sql[idx:], "SELECT") {
		t.Fatalf("expected ignorable content to be skipped, got %q", sql[idx:])
	}
}

func TestSkipWhitespace(t *testing.T) {
	t.Parallel()

	data := []byte("\t\n  VALUE")
	idx := skipWhitespace(data, 0)
	if idx != 4 {
		t.Fatalf("expected index 4, got %d", idx)
	}
}

func TestConsumeKeyword(t *testing.T) {
	t.Parallel()

	if next, ok := consumeKeyword([]byte("  IF NOT EXISTS"), 0, "IF"); !ok || next <= 0 {
		t.Fatalf("expected IF keyword to be consumed, ok=%v next=%d", ok, next)
	}

	if _, ok := consumeKeyword([]byte("INFORM"), 0, "IN"); ok {
		t.Fatal("did not expect partial keyword match")
	}
}

func TestConsumeIdentifier(t *testing.T) {
	t.Parallel()

	if next, ok := consumeIdentifier([]byte(`  "My"."Table"`), 0); !ok || next == 0 {
		t.Fatalf("expected quoted identifier to be consumed, ok=%v next=%d", ok, next)
	}

	if _, ok := consumeIdentifier([]byte("   ,"), 0); ok {
		t.Fatal("did not expect identifier before comma")
	}
}

func TestSkipParentheticalSection(t *testing.T) {
	t.Parallel()

	data := []byte("(SELECT (1 + 2)) trailing")
	end, ok := skipParentheticalSection(data, 0)
	if !ok || end != len("(SELECT (1 + 2))")-1 {
		t.Fatalf("expected parenthetical section to end at %d, got %d", len("(SELECT (1 + 2))")-1, end)
	}
}

func TestHasPrefixFold(t *testing.T) {
	t.Parallel()

	if !hasPrefixFold([]byte("Select"), "select") {
		t.Fatal("expected case-insensitive match")
	}

	if hasPrefixFold([]byte("Se"), "SELECT") {
		t.Fatal("did not expect prefix match when shorter than keyword")
	}
}

func TestParseStatementsReportsUnterminatedError(t *testing.T) {
	t.Parallel()

	_, err := parseStatements([]byte("CREATE TABLE widgets (id INT"))
	if err == nil || !strings.Contains(err.Error(), "unterminated") {
		t.Fatalf("expected unterminated statement error, got %v", err)
	}
}

func TestLocateMigrationsDir(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(filepath.Dir(wd)); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	}()

	dir, err := locateMigrationsDir()
	if err != nil {
		t.Fatalf("locate migrations dir: %v", err)
	}

	if !strings.HasSuffix(filepath.Clean(dir), filepath.Clean(migrationsRelativeDir)) {
		t.Fatalf("expected migrations directory suffix %q, got %q", migrationsRelativeDir, dir)
	}
}
