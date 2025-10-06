package sqlseed

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/oullin/database"
	"github.com/oullin/metal/env"
)

func SeedFromFile(conn *database.Connection, environment *env.Environment, filePath string) error {
	cleanedPath, err := validateFilePath(filePath)
	if err != nil {
		return err
	}

	fileContents, err := readSQLFile(cleanedPath)
	if err != nil {
		return err
	}

	if conn == nil {
		return errors.New("sqlseed: database connection is required")
	}

	if environment == nil {
		return errors.New("sqlseed: environment is required")
	}

	statements, err := parseStatements(fileContents)
	if err != nil {
		return err
	}

	if len(statements) == 0 {
		return errors.New("sqlseed: SQL file did not contain any executable statements")
	}

	ctx := context.Background()

	if err := prepareDatabase(ctx, conn, environment); err != nil {
		return err
	}

	return executeStatements(ctx, conn, statements)
}

const storageSQLDir = "storage/sql"
const migrationsDir = "database/infra/migrations"

func prepareDatabase(ctx context.Context, conn *database.Connection, environment *env.Environment) error {
	truncate := database.MakeTruncate(conn, environment)

	if err := truncate.Execute(); err != nil {
		return fmt.Errorf("sqlseed: truncate database: %w", err)
	}

	if err := runMigrations(ctx, conn); err != nil {
		return err
	}

	return nil
}

func runMigrations(ctx context.Context, conn *database.Connection) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("sqlseed: migrations directory %s not found", migrationsDir)
		}
		return fmt.Errorf("sqlseed: read migrations directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".up.sql") {
			files = append(files, filepath.Join(migrationsDir, name))
		}
	}

	sort.Strings(files)

	for _, path := range files {
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("sqlseed: read migration %s: %w", filepath.Base(path), err)
		}

		statements, err := parseStatements(contents)
		if err != nil {
			return fmt.Errorf("sqlseed: parse migration %s: %w", filepath.Base(path), err)
		}

		if len(statements) == 0 {
			continue
		}

		if err := executeStatements(ctx, conn, statements); err != nil {
			return fmt.Errorf("sqlseed: execute migration %s: %w", filepath.Base(path), err)
		}
	}

	return nil
}

func validateFilePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("sqlseed: file path is required")
	}

	if filepath.IsAbs(path) {
		return "", errors.New("sqlseed: absolute file paths are not supported")
	}

	cleanedInput := filepath.Clean(path)
	if ext := strings.ToLower(filepath.Ext(cleanedInput)); ext != ".sql" {
		return "", fmt.Errorf("sqlseed: unsupported file extension %q", filepath.Ext(cleanedInput))
	}

	base := filepath.Clean(storageSQLDir)
	baseWithSep := base + string(os.PathSeparator)

	var resolved string
	if strings.HasPrefix(cleanedInput, baseWithSep) {
		resolved = cleanedInput
	} else {
		resolved = filepath.Join(base, cleanedInput)
	}

	resolved = filepath.Clean(resolved)
	if !strings.HasPrefix(resolved, baseWithSep) {
		return "", fmt.Errorf("sqlseed: file path must be within %s", base)
	}

	return resolved, nil
}

func readSQLFile(path string) ([]byte, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("sqlseed: read file: %w", err)
	}

	if !utf8.Valid(contents) {
		return nil, fmt.Errorf("sqlseed: file %s contains non-UTF-8 data; ensure dumps are exported as plain text", path)
	}

	if len(bytes.TrimSpace(contents)) == 0 {
		return nil, fmt.Errorf("sqlseed: file %s is empty", path)
	}

	return contents, nil
}

type statement struct {
	sql      string
	copyData []byte
	isCopy   bool
}

func parseStatements(contents []byte) ([]statement, error) {
	var stmts []statement

	data := bytes.TrimSpace(contents)
	if len(data) == 0 {
		return nil, nil
	}

	leadingWhitespace := len(contents) - len(bytes.TrimLeftFunc(contents, func(r rune) bool {
		return unicode.IsSpace(r)
	}))

	var (
		idx            int
		start          int
		inSingleQuote  bool
		inDoubleQuote  bool
		inLineComment  bool
		inBlockComment bool
		dollarTag      string
	)

	start = skipIgnorableSections(data, start)
	idx = start

	for idx < len(data) {
		b := data[idx]

		switch {
		case inLineComment:
			if b == '\n' {
				inLineComment = false
			}
			idx++
			continue
		case inBlockComment:
			if b == '*' && idx+1 < len(data) && data[idx+1] == '/' {
				inBlockComment = false
				idx += 2
				continue
			}
			idx++
			continue
		case dollarTag != "":
			if b == '$' && hasDollarTag(data[idx:], dollarTag) {
				idx += len(dollarTag) + 2
				dollarTag = ""
				continue
			}
			idx++
			continue
		case inSingleQuote:
			if b == '\\' && idx+1 < len(data) {
				idx += 2
				continue
			}
			if b == '\'' {
				inSingleQuote = false
			}
			idx++
			continue
		case inDoubleQuote:
			if b == '"' {
				inDoubleQuote = false
			}
			idx++
			continue
		}

		if b == '-' && idx+1 < len(data) && data[idx+1] == '-' {
			inLineComment = true
			idx += 2
			continue
		}

		if b == '/' && idx+1 < len(data) && data[idx+1] == '*' {
			inBlockComment = true
			idx += 2
			continue
		}

		if b == '$' {
			tag, ok := readDollarTag(data[idx:])
			if ok {
				dollarTag = tag
				idx += len(tag) + 2
				continue
			}
		}

		if b == '\'' {
			inSingleQuote = true
			idx++
			continue
		}

		if b == '"' {
			inDoubleQuote = true
			idx++
			continue
		}

		if b != ';' {
			idx++
			continue
		}

		rawStmt := bytes.TrimSpace(data[start : idx+1])
		idx++
		if len(rawStmt) == 0 {
			start = skipIgnorableSections(data, idx)
			idx = start
			continue
		}

		trimmed := strings.TrimSpace(string(rawStmt))
		if isCopyFromStdin(trimmed) {
			copyStart := skipIgnorableSections(data, idx)
			copyLen, advance, err := extractCopyData(data[copyStart:])
			if err != nil {
				return nil, err
			}

			stmt := statement{
				sql:      strings.TrimSpace(strings.TrimSuffix(trimmed, ";")),
				copyData: append([]byte(nil), data[copyStart:copyStart+copyLen]...),
				isCopy:   true,
			}
			stmts = append(stmts, stmt)

			idx = copyStart + advance
			start = skipIgnorableSections(data, idx)
			idx = start
			continue
		}

		stmt := statement{sql: strings.TrimSpace(strings.TrimSuffix(trimmed, ";"))}
		stmts = append(stmts, stmt)

		start = skipIgnorableSections(data, idx)
		idx = start
	}

	start = skipIgnorableSections(data, start)
	if start < len(data) {
		if len(bytes.TrimSpace(data[start:])) != 0 {
			originalIdx := start + leadingWhitespace
			line, column := lineAndColumn(contents, originalIdx)
			preview := formatSnippet(data[start:])
			return nil, fmt.Errorf("sqlseed: SQL file ended with an unterminated statement at line %d, column %d near %q", line, column, preview)
		}
	}

	return stmts, nil
}

func skipIgnorableSections(data []byte, idx int) int {
	for idx < len(data) {
		r, size := utf8DecodeRune(data[idx:])
		if size == 0 {
			return idx
		}
		switch r {
		case ' ', '\n', '\r', '\t':
			idx += size
			continue
		}

		if data[idx] == '-' && idx+1 < len(data) && data[idx+1] == '-' {
			idx += 2
			for idx < len(data) {
				if data[idx] == '\n' || data[idx] == '\r' {
					idx++
					break
				}
				idx++
			}
			continue
		}

		if data[idx] == '/' && idx+1 < len(data) && data[idx+1] == '*' {
			idx += 2
			for idx < len(data) {
				if data[idx] == '*' && idx+1 < len(data) && data[idx+1] == '/' {
					idx += 2
					break
				}
				idx++
			}
			continue
		}

		return idx
	}

	return idx
}

func utf8DecodeRune(data []byte) (rune, int) {
	if len(data) == 0 {
		return 0, 0
	}
	r, size := utf8.DecodeRune(data)
	if r == utf8.RuneError && size == 1 {
		return rune(data[0]), 1
	}
	return r, size
}

func readDollarTag(data []byte) (string, bool) {
	if len(data) < 2 || data[0] != '$' {
		return "", false
	}

	end := 1
	for end < len(data) {
		c := data[end]
		if c == '$' {
			return string(data[1:end]), true
		}
		if !isDollarTagChar(c) {
			return "", false
		}
		end++
	}

	return "", false
}

func isDollarTagChar(b byte) bool {
	return b == '_' || b == '$' || unicode.IsLetter(rune(b)) || unicode.IsDigit(rune(b))
}

func hasDollarTag(data []byte, tag string) bool {
	marker := "$" + tag + "$"
	return len(data) >= len(marker) && string(data[:len(marker)]) == marker
}

func isCopyFromStdin(stmt string) bool {
	upper := strings.ToUpper(stmt)
	return strings.HasPrefix(upper, "COPY ") && strings.Contains(upper, "FROM STDIN")
}

func extractCopyData(data []byte) (int, int, error) {
	patterns := []struct {
		marker  []byte
		include int
	}{
		{[]byte("\r\n\\.\r\n"), 2},
		{[]byte("\n\\.\n"), 1},
		{[]byte("\n\\.\r\n"), 1},
		{[]byte("\r\n\\.\n"), 2},
		{[]byte("\\.\r\n"), 0},
		{[]byte("\\.\n"), 0},
		{[]byte("\\."), 0},
	}

	for _, pattern := range patterns {
		if idx := bytes.Index(data, pattern.marker); idx != -1 {
			copyEnd := idx + pattern.include
			advance := idx + len(pattern.marker)
			return copyEnd, advance, nil
		}
	}

	return 0, 0, errors.New("sqlseed: COPY statement missing terminator")
}

func executeStatements(ctx context.Context, conn *database.Connection, statements []statement) error {
	sqlDB, err := conn.Sql().DB()
	if err != nil {
		return fmt.Errorf("sqlseed: retrieve sql db: %w", err)
	}

	sqlConn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("sqlseed: acquire connection: %w", err)
	}
	defer sqlConn.Close()

	var execErr error
	err = sqlConn.Raw(func(driverConn interface{}) error {
		stdlibConn, ok := driverConn.(*stdlib.Conn)
		if !ok {
			return errors.New("sqlseed: unexpected driver connection type")
		}

		pgxConn := stdlibConn.Conn()
		tx, err := pgxConn.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return fmt.Errorf("sqlseed: begin transaction: %w", err)
		}

		committed := false
		defer func() {
			if committed {
				return
			}
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				execErr = errors.Join(execErr, fmt.Errorf("sqlseed: rollback failed: %w", rbErr))
			}
		}()

		for idx, stmt := range statements {
			statementNumber := idx + 1
			preview := formatSnippet([]byte(stmt.sql))
			if stmt.isCopy {
				if err := executeCopy(ctx, tx.Conn().PgConn(), stmt); err != nil {
					execErr = fmt.Errorf("sqlseed: executing COPY statement %d near %q failed: %w", statementNumber, preview, err)
					return execErr
				}
				continue
			}

			if _, err := tx.Exec(ctx, stmt.sql); err != nil {
				if skip, reason := shouldSkipExecError(stmt, err); skip {
					fmt.Fprintf(os.Stderr, "sqlseed: skipped statement %d near %q: %s\n", statementNumber, preview, reason)
					continue
				}
				execErr = fmt.Errorf("sqlseed: executing SQL statement %d near %q failed: %w", statementNumber, preview, err)
				return execErr
			}
		}

		if err := tx.Commit(ctx); err != nil {
			execErr = fmt.Errorf("sqlseed: commit failed: %w", err)
			return execErr
		}

		committed = true
		return nil
	})
	if err != nil {
		if execErr != nil {
			return execErr
		}
		return err
	}

	if execErr != nil {
		return execErr
	}

	return nil
}

func executeCopy(ctx context.Context, pgConn *pgconn.PgConn, stmt statement) error {
	reader := bytes.NewReader(stmt.copyData)
	sql := strings.TrimSpace(stmt.sql)
	sql = strings.TrimSuffix(sql, ";")
	if _, err := pgConn.CopyFrom(ctx, reader, sql); err != nil {
		return fmt.Errorf("sqlseed: executing COPY failed: %w", err)
	}
	return nil
}

func shouldSkipExecError(stmt statement, err error) (bool, string) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false, ""
	}

	upper := strings.ToUpper(strings.TrimSpace(stmt.sql))

	switch pgErr.Code {
	case "42P07", "42P06", "42710":
		if strings.HasPrefix(upper, "CREATE ") {
			return true, fmt.Sprintf("object already exists (%s)", pgErr.Message)
		}
	case "42704":
		if strings.Contains(upper, " OWNER TO ") {
			return true, fmt.Sprintf("owner skipped (%s)", pgErr.Message)
		}
	}

	return false, ""
}

func formatSnippet(data []byte) string {
	const maxRunes = 120

	snippet := strings.TrimSpace(string(data))
	if snippet == "" {
		return "<empty>"
	}

	replacer := strings.NewReplacer("\r", " ", "\n", " ", "\t", " ")
	snippet = replacer.Replace(snippet)
	snippet = strings.Join(strings.Fields(snippet), " ")

	runes := []rune(snippet)
	if len(runes) > maxRunes {
		snippet = string(runes[:maxRunes-1]) + "…"
	}

	return snippet
}

func lineAndColumn(src []byte, index int) (int, int) {
	line := 1
	column := 1

	if index < 0 {
		index = 0
	}
	if index > len(src) {
		index = len(src)
	}

	i := 0
	for i < index {
		b := src[i]
		if b == '\r' {
			line++
			column = 1
			if i+1 < len(src) && src[i+1] == '\n' {
				i += 2
				continue
			}
			i++
			continue
		}
		if b == '\n' {
			line++
			column = 1
			i++
			continue
		}

		column++
		i++
	}

	return line, column
}
