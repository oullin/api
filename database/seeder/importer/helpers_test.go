package importer

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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
			err:       &fakeSQLStateError{code: "23505", msg: "duplicate key value violates unique constraint \"schema_migrations_pkey\""},
			wantSkip:  true,
			wantMatch: "duplicate migration row",
		},
		{
			name:      "primary key already exists",
			sql:       "ALTER TABLE public.widgets ADD CONSTRAINT widgets_pkey PRIMARY KEY (id);",
			err:       &fakeSQLStateError{code: "42P16", msg: "multiple primary keys for table are not allowed"},
			wantSkip:  true,
			wantMatch: "primary key",
		},
		{
			name:      "owner change missing role",
			sql:       "ALTER TABLE public.widgets OWNER TO missing_role;",
			err:       &fakeSQLStateError{code: "42704", msg: "role \"missing_role\" does not exist"},
			wantSkip:  true,
			wantMatch: "owner skipped",
		},
		{
			name:      "object already exists",
			sql:       "CREATE TABLE public.widgets (id INT);",
			err:       &fakeSQLStateError{code: "42P07", msg: "relation \"widgets\" already exists"},
			wantSkip:  true,
			wantMatch: "object already exists",
		},
		{
			name:      "missing relation",
			sql:       "DROP TABLE public.widgets;",
			err:       &fakeSQLStateError{code: "42P01", msg: "relation \"widgets\" does not exist"},
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

func TestDecodeCopyRowsParsesData(t *testing.T) {
	t.Parallel()

	rows, err := decodeCopyRows([]byte("1\talpha\n2\t\\N"), 2)
	if err != nil {
		t.Fatalf("decodeCopyRows unexpected error: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}

	if rows[0][0] != "1" || rows[0][1] != "alpha" {
		t.Fatalf("unexpected first row: %+v", rows[0])
	}

	if rows[1][0] != "2" || rows[1][1] != nil {
		t.Fatalf("expected second row to contain NULL value, got %+v", rows[1])
	}
}

func TestDecodeCopyRowsDetectsMismatch(t *testing.T) {
	t.Parallel()

	_, err := decodeCopyRows([]byte("1\talpha\n2"), 2)
	if err == nil || !strings.Contains(err.Error(), "row 2") {
		t.Fatalf("expected mismatch error for row 2, got %v", err)
	}
}

func TestDecodeCopyFieldHandlesEscapes(t *testing.T) {
	t.Parallel()

	field := "line\\nwith\\ttext\\\\and\\101"
	value, err := decodeCopyField(field)
	if err != nil {
		t.Fatalf("decodeCopyField unexpected error: %v", err)
	}

	if value != "line\nwith\ttext\\andA" {
		t.Fatalf("unexpected decoded field %q", value)
	}
}

func TestDecodeCopyFieldErrorsOnTrailingEscape(t *testing.T) {
	t.Parallel()

	if _, err := decodeCopyField("abc\\"); err == nil || !strings.Contains(err.Error(), "escape prefix") {
		t.Fatalf("expected trailing escape error, got %v", err)
	}
}

func TestParseCopyStatementParsesTableAndColumns(t *testing.T) {
	t.Parallel()

	table, columns, err := parseCopyStatement("COPY widgets (id, name) FROM STDIN")
	if err != nil {
		t.Fatalf("parseCopyStatement unexpected error: %v", err)
	}

	if table != "widgets" {
		t.Fatalf("unexpected table %q", table)
	}

	want := []string{"id", "name"}
	if !reflect.DeepEqual(columns, want) {
		t.Fatalf("unexpected columns %#v", columns)
	}
}

func TestParseCopyStatementValidatesInput(t *testing.T) {
	t.Parallel()

	cases := []string{
		"",
		"INSERT INTO widgets",
		"COPY FROM STDIN",
		"COPY widgets (id FROM STDIN",
		"COPY widgets FROM file",
	}

	for _, input := range cases {
		if _, _, err := parseCopyStatement(input); err == nil {
			t.Fatalf("expected error for input %q", input)
		}
	}
}

func TestParseIdentifierListParsesQuotedValues(t *testing.T) {
	t.Parallel()

	list, err := parseIdentifierList(`id, "Name", slug`)
	if err != nil {
		t.Fatalf("parseIdentifierList unexpected error: %v", err)
	}

	want := []string{"id", `"Name"`, "slug"}
	if !reflect.DeepEqual(list, want) {
		t.Fatalf("unexpected identifier list %#v", list)
	}

	if _, err := parseIdentifierList(`"unterminated`); err == nil {
		t.Fatal("expected error for unterminated quotes")
	}
}

func TestSplitTableName(t *testing.T) {
	t.Parallel()

	schema, name := splitTableName(`"Public"."Widgets"`)
	if schema != "Public" || name != "Widgets" {
		t.Fatalf("unexpected split result schema=%q name=%q", schema, name)
	}

	schema, name = splitTableName("widgets")
	if schema != "" || name != "widgets" {
		t.Fatalf("expected empty schema for bare identifier, got schema=%q name=%q", schema, name)
	}
}

func TestSplitQualifiedIdentifierParts(t *testing.T) {
	t.Parallel()

	parts := splitQualifiedIdentifierParts(`"Public"."Widgets".child`)
	want := []string{`"Public"`, `"Widgets"`, "child"}
	if !reflect.DeepEqual(parts, want) {
		t.Fatalf("unexpected parts %#v", parts)
	}

	if len(splitQualifiedIdentifierParts("")) != 0 {
		t.Fatal("expected empty slice for empty identifier")
	}
}

func TestQuoteIdentifier(t *testing.T) {
	t.Parallel()

	if quoteIdentifier("widgets") != "widgets" {
		t.Fatal("expected simple identifier to remain unquoted")
	}

	if quoteIdentifier("Widgets") != `"Widgets"` {
		t.Fatal("expected mixed case identifier to be quoted")
	}

	if quoteIdentifier("needs space") != `"needs space"` {
		t.Fatal("expected identifier with space to be quoted")
	}
}

func TestSQLStateFromErrorPrefersInterface(t *testing.T) {
	t.Parallel()

	code, message := sqlStateFromError(&fakeSQLStateError{code: "23505", msg: "duplicate"})
	if code != "23505" || message != "duplicate" {
		t.Fatalf("unexpected code/message %q %q", code, message)
	}
}

func TestSQLStateFromErrorParsesMessage(t *testing.T) {
	t.Parallel()

	code, _ := sqlStateFromError(errors.New("ERROR: boom (SQLSTATE 42P01)"))
	if code != "42P01" {
		t.Fatalf("expected SQLSTATE 42P01, got %q", code)
	}

	code, _ = sqlStateFromError(errors.New("other error"))
	if code != "" {
		t.Fatalf("expected empty SQLSTATE, got %q", code)
	}
}

func TestIsSafeToIgnoreRollbackError(t *testing.T) {
	t.Parallel()

	if !isSafeToIgnoreRollbackError(errors.New("transaction has already been committed")) {
		t.Fatal("expected committed message to be ignorable")
	}

	if isSafeToIgnoreRollbackError(errors.New("different error")) {
		t.Fatal("did not expect unrelated error to be ignorable")
	}
}

func TestUtf8DecodeLastRune(t *testing.T) {
	t.Parallel()

	if r, size := utf8DecodeLastRune([]byte("é")); r != 'é' || size != 2 {
		t.Fatalf("expected last rune é size 2, got %q size %d", r, size)
	}

	if r, size := utf8DecodeLastRune([]byte{0xff}); r != rune(0xff) || size != 1 {
		t.Fatalf("expected fallback rune 0xff size 1, got %q size %d", r, size)
	}
}

func TestLocateStatementStartSkipsComments(t *testing.T) {
	t.Parallel()

	sql := "-- comment\n/* block */\nSELECT 1"
	if out := locateStatementStart(sql); !strings.HasPrefix(out, "SELECT 1") {
		t.Fatalf("expected SELECT statement, got %q", out)
	}

	sql = "$$do$$ SELECT 1$$do$$; INSERT INTO foo VALUES (1);"
	if out := locateStatementStart(sql); !strings.HasPrefix(out, "SELECT 1$$do$$") {
		t.Fatalf("expected dollar-quoted block, got %q", out)
	}
}

func TestSkipParentheticalSectionDetectsUnbalanced(t *testing.T) {
	t.Parallel()

	if _, ok := skipParentheticalSection([]byte("(SELECT 1"), 0); ok {
		t.Fatal("expected unbalanced parentheses to return false")
	}
}

func TestSkipParentheticalSectionHandlesQuotes(t *testing.T) {
	t.Parallel()

	data := []byte("(SELECT '(())'::text, 1)")
	end, ok := skipParentheticalSection(data, 0)
	if !ok || end != len(data)-1 {
		t.Fatalf("expected section to end at %d, got end=%d ok=%v", len(data)-1, end, ok)
	}
}

func TestParseStatementsHandlesWindowsNewlines(t *testing.T) {
	t.Parallel()

	sql := "INSERT INTO foo VALUES (1);\r\n-- comment\r\nINSERT INTO foo VALUES (2);\r\n"
	stmts, err := parseStatements([]byte(sql))
	if err != nil {
		t.Fatalf("parseStatements unexpected error: %v", err)
	}

	if len(stmts) != 2 {
		t.Fatalf("expected two statements, got %d", len(stmts))
	}

	if stmts[1].sql != "INSERT INTO foo VALUES (2)" {
		t.Fatalf("unexpected second statement %q", stmts[1].sql)
	}
}

func TestSkipLeadingCTESectionsWithMultipleCTEs(t *testing.T) {
	t.Parallel()

	sql := "WITH a AS (SELECT 1), b AS (SELECT 2) INSERT INTO foo VALUES (1);"
	idx := skipLeadingCTESections([]byte(sql), 0)
	if !strings.HasPrefix(sql[idx:], "INSERT") {
		t.Fatalf("expected cursor at INSERT, got %q", sql[idx:])
	}
}

func TestSkipLeadingCTESectionsWithRecursiveCTE(t *testing.T) {
	t.Parallel()

	sql := "WITH RECURSIVE t AS (SELECT 1) SELECT * FROM t;"
	idx := skipLeadingCTESections([]byte(sql), 0)
	if !strings.HasPrefix(sql[idx:], "SELECT") {
		t.Fatalf("expected SELECT start, got %q", sql[idx:])
	}
}

func TestSkipLeadingCTESectionsIgnoresKeywordInString(t *testing.T) {
	t.Parallel()

	sql := "SELECT 'WITH clause' AS note;"
	if idx := skipLeadingCTESections([]byte(sql), 0); idx != 0 {
		t.Fatalf("expected index 0 for non-CTE, got %d", idx)
	}
}

func TestSkipParentheticalSectionRequiresOpeningParen(t *testing.T) {
	t.Parallel()

	if _, ok := skipParentheticalSection([]byte("SELECT (1)"), 0); ok {
		t.Fatal("expected false when starting index is not a parenthesis")
	}
}

func TestUtf8DecodeRuneFallback(t *testing.T) {
	t.Parallel()

	if r, size := utf8DecodeRune([]byte{0xff}); r != rune(0xff) || size != 1 {
		t.Fatalf("expected fallback rune 0xff size 1, got %q size %d", r, size)
	}
}

func TestParseStatementsErrorsWhenCopyTerminatorMissing(t *testing.T) {
	t.Parallel()

	sql := "COPY foo (id) FROM stdin;\n1\n"
	if _, err := parseStatements([]byte(sql)); err == nil || !strings.Contains(err.Error(), "missing terminator") {
		t.Fatalf("expected terminator error, got %v", err)
	}
}

func TestParseStatementsHandlesDollarQuotedStrings(t *testing.T) {
	t.Parallel()

	sql := "CREATE FUNCTION test() RETURNS void AS $$BEGIN PERFORM 1; END$$ LANGUAGE plpgsql;"
	stmts, err := parseStatements([]byte(sql))
	if err != nil {
		t.Fatalf("parseStatements unexpected error: %v", err)
	}

	if len(stmts) == 0 {
		t.Fatal("expected at least one statement")
	}

	if !strings.HasPrefix(stmts[0].sql, "CREATE FUNCTION test() RETURNS void AS $$BEGIN") {
		t.Fatalf("unexpected function statement %q", stmts[0].sql)
	}
}

func TestStatementTargetHandlesCreateUniqueIndex(t *testing.T) {
	t.Parallel()

	table, op := statementTarget("CREATE UNIQUE INDEX idx ON public.widgets (id);")
	if table != "public.widgets" || op != "CREATE INDEX" {
		t.Fatalf("unexpected target %q operation %q", table, op)
	}
}

func TestStatementTargetHandlesAlterSequence(t *testing.T) {
	t.Parallel()

	table, op := statementTarget("ALTER SEQUENCE public.widgets_id_seq OWNED BY public.widgets.id;")
	if table != "public.widgets_id_seq" || op != "ALTER SEQUENCE" {
		t.Fatalf("unexpected target %q operation %q", table, op)
	}
}

func TestStatementTargetHandlesDropIndex(t *testing.T) {
	t.Parallel()

	table, op := statementTarget("DROP INDEX IF EXISTS widgets_idx;")
	if table != "widgets_idx" || op != "DROP INDEX" {
		t.Fatalf("unexpected drop index target %q operation %q", table, op)
	}
}

func TestSkipParentheticalSectionHandlesComments(t *testing.T) {
	t.Parallel()

	data := []byte("(SELECT 1 /* comment ) */ -- line\n)")
	end, ok := skipParentheticalSection(data, 0)
	if !ok || end != len(data)-1 {
		t.Fatalf("expected section to end at %d, got end=%d ok=%v", len(data)-1, end, ok)
	}
}

func TestParseStatementsSkipsConnectCommands(t *testing.T) {
	t.Parallel()

	sql := "\\connect database\nINSERT INTO foo VALUES (1);"
	stmts, err := parseStatements([]byte(sql))
	if err != nil {
		t.Fatalf("parseStatements unexpected error: %v", err)
	}

	if len(stmts) != 1 || stmts[0].sql != "INSERT INTO foo VALUES (1)" {
		t.Fatalf("expected single insert statement, got %#v", stmts)
	}
}

func TestParseStatementsHandlesQuotedIdentifiers(t *testing.T) {
	t.Parallel()

	sql := `INSERT INTO "My"."Table" ("Col") VALUES ('value; still string');`
	stmts, err := parseStatements([]byte(sql))
	if err != nil {
		t.Fatalf("parseStatements unexpected error: %v", err)
	}

	if len(stmts) != 1 || !strings.Contains(stmts[0].sql, `"My"."Table"`) {
		t.Fatalf("unexpected parsed statement %#v", stmts)
	}
}

func TestLocateStatementStartWithoutKeyword(t *testing.T) {
	t.Parallel()

	if out := locateStatementStart("   "); out != "   " {
		t.Fatalf("expected original whitespace, got %q", out)
	}
}

type fakeSQLStateError struct {
	code string
	msg  string
}

func (e *fakeSQLStateError) Error() string {
	return e.msg
}

func (e *fakeSQLStateError) SQLState() string {
	return e.code
}
