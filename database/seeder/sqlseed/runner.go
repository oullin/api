package sqlseed

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/oullin/database"
)

func SeedFromFile(conn *database.Connection, filePath string) error {
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

	statements, err := parseStatements(fileContents)
	if err != nil {
		return err
	}

	if len(statements) == 0 {
		return errors.New("sqlseed: SQL file did not contain any executable statements")
	}

	return executeStatements(context.Background(), conn, statements)
}

const storageSQLDir = "storage/sql"

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
