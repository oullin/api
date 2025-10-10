package importer

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

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
		return errors.New("importer: database connection is required")
	}

	if environment == nil {
		return errors.New("importer: environment is required")
	}

	statements, err := parseStatements(fileContents)
	if err != nil {
		return err
	}

	if len(statements) == 0 {
		return errors.New("importer: SQL file did not contain any executable statements")
	}

	ctx := context.Background()

	if err := prepareDatabase(ctx, conn, environment); err != nil {
		return err
	}

	return executeStatements(ctx, conn, statements, executeOptions{
		disableConstraints: true,
		skipTables:         excludedSeedTables,
	})
}

func prepareDatabase(ctx context.Context, conn *database.Connection, environment *env.Environment) error {
	truncate := database.MakeTruncate(conn, environment)

	if err := truncate.Execute(); err != nil {
		return fmt.Errorf("importer: truncate database: %w", err)
	}

	if err := runMigrations(ctx, conn); err != nil {
		return err
	}

	return nil
}

func runMigrations(ctx context.Context, conn *database.Connection) error {
	dir, err := locateMigrationsDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("importer: read migrations directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".up.sql") {
			files = append(files, filepath.Join(dir, name))
		}
	}

	sort.Strings(files)

	for _, path := range files {
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("importer: read migration %s: %w", filepath.Base(path), err)
		}

		statements, err := parseStatements(contents)
		if err != nil {
			return fmt.Errorf("importer: parse migration %s: %w", filepath.Base(path), err)
		}

		if len(statements) == 0 {
			continue
		}

		if err := executeStatements(ctx, conn, statements, executeOptions{}); err != nil {
			return fmt.Errorf("importer: execute migration %s: %w", filepath.Base(path), err)
		}
	}

	return nil
}

func locateMigrationsDir() (string, error) {
	cleaned := filepath.Clean(migrationsRelativeDir)

	if info, err := os.Stat(cleaned); err == nil && info.IsDir() {
		return cleaned, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("importer: determine working directory: %w", err)
	}

	dir := wd
	for {
		candidate := filepath.Join(dir, cleaned)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("importer: migrations directory %s not found", cleaned)
}

func validateFilePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("importer: file path is required")
	}

	if filepath.IsAbs(path) {
		return "", errors.New("importer: absolute file paths are not supported")
	}

	cleanedInput := filepath.Clean(path)
	if ext := strings.ToLower(filepath.Ext(cleanedInput)); ext != ".sql" {
		return "", fmt.Errorf("importer: unsupported file extension %q", filepath.Ext(cleanedInput))
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
		return "", fmt.Errorf("importer: file path must be within %s", base)
	}

	return resolved, nil
}

func readSQLFile(path string) ([]byte, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("importer: read file: %w", err)
	}

	if !utf8.Valid(contents) {
		return nil, fmt.Errorf("importer: file %s contains non-UTF-8 data; ensure dumps are exported as plain text", path)
	}

	if len(bytes.TrimSpace(contents)) == 0 {
		return nil, fmt.Errorf("importer: file %s is empty", path)
	}

	return contents, nil
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
			copyStart := skipWhitespace(data, idx)
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
			return nil, fmt.Errorf("importer: SQL file ended with an unterminated statement at line %d, column %d near %q", line, column, preview)
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

		if data[idx] == '\\' {
			idx++
			for idx < len(data) && data[idx] != '\n' && data[idx] != '\r' {
				idx++
			}

			if idx < len(data) {
				if data[idx] == '\r' {
					idx++
					if idx < len(data) && data[idx] == '\n' {
						idx++
					}
				} else if data[idx] == '\n' {
					idx++
				}
			}

			continue
		}

		return idx
	}

	return idx
}

func skipWhitespace(data []byte, idx int) int {
	for idx < len(data) {
		r, size := utf8DecodeRune(data[idx:])
		if size == 0 {
			break
		}
		if unicode.IsSpace(r) {
			idx += size
			continue
		}
		break
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

func utf8DecodeLastRune(data []byte) (rune, int) {
	if len(data) == 0 {
		return 0, 0
	}
	r, size := utf8.DecodeLastRune(data)
	if r == utf8.RuneError && size == 1 {
		return rune(data[len(data)-1]), 1
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
	if len(data) == 0 {
		return 0, 0, errors.New("importer: COPY statement missing terminator")
	}

	offset := 0
	for offset < len(data) {
		lineEndRelative := bytes.IndexByte(data[offset:], '\n')
		if lineEndRelative == -1 {
			line := data[offset:]
			trimmed := bytes.TrimSuffix(line, []byte{'\r'})
			if bytes.Equal(trimmed, []byte("\\.")) {
				copyEnd := offset
				advance := len(data)
				return copyEnd, advance, nil
			}

			break
		}

		lineEnd := offset + lineEndRelative
		line := data[offset:lineEnd]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		if bytes.Equal(line, []byte("\\.")) {
			copyEnd := offset
			advance := lineEnd + 1
			return copyEnd, advance, nil
		}

		offset = lineEnd + 1
	}

	return 0, 0, errors.New("importer: COPY statement missing terminator")
}

func executeStatements(ctx context.Context, conn *database.Connection, statements []statement, opts executeOptions) error {
	sqlDB, err := conn.Sql().DB()
	if err != nil {
		return fmt.Errorf("importer: retrieve sql db: %w", err)
	}

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("importer: begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := tx.Rollback(); rbErr != nil && !isSafeToIgnoreRollbackError(rbErr) {
			fmt.Fprintf(os.Stderr, "importer: rollback failed: %v\n", rbErr)
		}
	}()

	if opts.disableConstraints {
		if _, err := tx.ExecContext(ctx, "SET LOCAL session_replication_role = 'replica'"); err != nil {
			return fmt.Errorf("importer: disable constraints failed: %w", err)
		}
	}

	for idx, stmt := range statements {
		statementNumber := idx + 1
		preview := formatSnippet([]byte(stmt.sql))

		if skip, reason := shouldSkipStatement(stmt, opts.skipTables); skip {
			fmt.Fprintf(os.Stderr, "importer: skipped statement %d near %q: %s\n", statementNumber, preview, reason)
			continue
		}

		savepoint := fmt.Sprintf("importer_sp_%d", statementNumber)
		if _, err := tx.ExecContext(ctx, "SAVEPOINT "+savepoint); err != nil {
			return fmt.Errorf("importer: begin savepoint for statement %d near %q: %w", statementNumber, preview, err)
		}

		execErr := func() error {
			if stmt.isCopy {
				if err := executeCopy(ctx, tx, stmt); err != nil {
					return fmt.Errorf("importer: executing COPY statement %d near %q failed: %w", statementNumber, preview, err)
				}
				return nil
			}

			if _, err := tx.ExecContext(ctx, stmt.sql); err != nil {
				if skip, reason := shouldSkipExecError(stmt, err); skip {
					if rbErr := rollbackToSavepoint(ctx, tx, savepoint); rbErr != nil {
						return errors.Join(fmt.Errorf("importer: skipped statement %d near %q due to %s", statementNumber, preview, reason), rbErr)
					}
					fmt.Fprintf(os.Stderr, "importer: skipped statement %d near %q: %s\n", statementNumber, preview, reason)
					return nil
				}

				if rbErr := rollbackToSavepoint(ctx, tx, savepoint); rbErr != nil {
					return errors.Join(fmt.Errorf("importer: executing SQL statement %d near %q failed: %w", statementNumber, preview, err), rbErr)
				}

				return fmt.Errorf("importer: executing SQL statement %d near %q failed: %w", statementNumber, preview, err)
			}

			return nil
		}()

		if execErr != nil {
			return execErr
		}

		if _, err := tx.ExecContext(ctx, "RELEASE SAVEPOINT "+savepoint); err != nil {
			return fmt.Errorf("importer: release savepoint for statement %d near %q failed: %w", statementNumber, preview, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("importer: commit failed: %w", err)
	}

	committed = true
	return nil
}

func rollbackToSavepoint(ctx context.Context, tx *sql.Tx, name string) error {
	if _, err := tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+name); err != nil {
		return fmt.Errorf("importer: rollback savepoint %s failed: %w", name, err)
	}

	return nil
}

func executeCopy(ctx context.Context, tx *sql.Tx, stmt statement) error {
	table, columns, err := parseCopyStatement(stmt.sql)
	if err != nil {
		return err
	}

	if len(columns) == 0 {
		columns, err = resolveCopyColumns(ctx, tx, table)
		if err != nil {
			return err
		}
	}

	rows, err := decodeCopyRows(stmt.copyData, len(columns))
	if err != nil {
		return err
	}

	placeholder := make([]string, len(columns))
	for i := range placeholder {
		placeholder[i] = fmt.Sprintf("$%d", i+1)
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholder, ", "))

	for _, row := range rows {
		if _, err := tx.ExecContext(ctx, insertSQL, row...); err != nil {
			return fmt.Errorf("importer: insert from COPY failed: %w", err)
		}
	}

	return nil
}

func isSafeToIgnoreRollbackError(err error) bool {
	if err == nil {
		return true
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "transaction has already been committed") || strings.Contains(lower, "transaction has already been rolled back")
}

func shouldSkipExecError(stmt statement, err error) (bool, string) {
	code, message := sqlStateFromError(err)
	if code == "" {
		return false, ""
	}

	upper := strings.ToUpper(strings.TrimSpace(stmt.sql))

	switch code {
	case "42P07", "42P06", "42710":
		if strings.HasPrefix(upper, "CREATE ") || strings.HasPrefix(upper, "ALTER TABLE ") || strings.HasPrefix(upper, "ALTER INDEX ") {
			return true, fmt.Sprintf("object already exists (%s)", message)
		}
	case "42P16":
		if strings.HasPrefix(upper, "ALTER TABLE ") || strings.HasPrefix(upper, "CREATE TABLE ") {
			if strings.Contains(upper, "PRIMARY KEY") {
				return true, fmt.Sprintf("primary key skipped (%s)", message)
			}
		}
	case "23505":
		if strings.HasPrefix(upper, "INSERT INTO ") {
			target, _ := statementTarget(stmt.sql)
			normalized := normalizeQualifiedIdentifier(target)
			switch normalized {
			case "schema_migrations", "public.schema_migrations":
				return true, fmt.Sprintf("duplicate migration row skipped (%s)", message)
			}
		}
	case "42704":
		if strings.Contains(upper, " OWNER TO ") {
			return true, fmt.Sprintf("owner skipped (%s)", message)
		}
	case "42P01":
		if strings.HasPrefix(upper, "ALTER TABLE") || strings.HasPrefix(upper, "DROP ") {
			return true, fmt.Sprintf("relation skipped (%s)", message)
		}
	}

	return false, ""
}

func shouldSkipStatement(stmt statement, skipTables map[string]struct{}) (bool, string) {
	if len(skipTables) == 0 {
		return false, ""
	}

	target, operation := statementTarget(stmt.sql)
	if target == "" {
		return false, ""
	}

	if !shouldExcludeIdentifier(target, skipTables) {
		return false, ""
	}

	normalized := normalizeQualifiedIdentifier(target)
	if normalized == "" {
		normalized = target
	}

	if operation == "" {
		operation = "statement"
	}

	return true, fmt.Sprintf("%s targets excluded identifier %s", operation, normalized)
}

func parseCopyStatement(sql string) (string, []string, error) {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return "", nil, errors.New("importer: COPY statement is empty")
	}

	if !strings.HasPrefix(strings.ToUpper(trimmed), "COPY ") {
		return "", nil, errors.New("importer: COPY statement must begin with COPY")
	}

	remainder := strings.TrimSpace(trimmed[len("COPY "):])
	table := firstIdentifier(remainder, true)
	if table == "" {
		return "", nil, errors.New("importer: COPY statement missing table name")
	}

	remainder = strings.TrimSpace(remainder[len(table):])
	var columns []string
	if strings.HasPrefix(remainder, "(") {
		end, ok := skipParentheticalSection([]byte(remainder), 0)
		if !ok {
			return "", nil, errors.New("importer: COPY statement column list is malformed")
		}

		list := remainder[1:end]
		parsed, err := parseIdentifierList(list)
		if err != nil {
			return "", nil, err
		}
		columns = parsed

		remainder = strings.TrimSpace(remainder[end+1:])
	}

	upperRemainder := strings.ToUpper(remainder)
	idx := strings.Index(upperRemainder, "FROM")
	if idx == -1 {
		return "", nil, errors.New("importer: COPY statement missing FROM clause")
	}

	afterFrom := strings.TrimSpace(remainder[idx+len("FROM"):])
	if !strings.HasPrefix(strings.ToUpper(afterFrom), "STDIN") {
		return "", nil, errors.New("importer: COPY statement must target STDIN")
	}

	return table, columns, nil
}

func parseIdentifierList(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	var (
		parts    []string
		current  strings.Builder
		inQuotes bool
	)

	data := []byte(input)

	for i := 0; i < len(data); {
		r, size := utf8DecodeRune(data[i:])
		if size == 0 {
			break
		}

		switch r {
		case '"':
			current.WriteRune(r)
			if inQuotes && i+size < len(data) {
				next, nextSize := utf8DecodeRune(data[i+size:])
				if next == '"' {
					current.WriteRune(next)
					i += size + nextSize
					continue
				}
			}
			inQuotes = !inQuotes
			i += size
			continue
		case ',':
			if !inQuotes {
				part := strings.TrimSpace(current.String())
				if part != "" {
					parts = append(parts, part)
				}
				current.Reset()
				i += size
				continue
			}
		}

		current.WriteRune(r)
		i += size
	}

	if current.Len() > 0 {
		part := strings.TrimSpace(current.String())
		if part != "" {
			parts = append(parts, part)
		}
	}

	if inQuotes {
		return nil, errors.New("importer: COPY column list has unterminated quotes")
	}

	return parts, nil
}

const (
	copyScannerInitialCapacity = 64 * 1024
	copyScannerMaxCapacity     = 16 * 1024 * 1024
)

func decodeCopyRows(data []byte, columnCount int) ([][]any, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	maxCapacity := copyScannerMaxCapacity
	if maxCapacity < copyScannerInitialCapacity {
		maxCapacity = copyScannerInitialCapacity
	}
	if len(data) > maxCapacity {
		maxCapacity = len(data)
	}

	scanner.Buffer(make([]byte, copyScannerInitialCapacity), maxCapacity)
	var rows [][]any
	line := 0

	for scanner.Scan() {
		text := scanner.Text()
		line++

		fields := strings.Split(text, "\t")
		if columnCount == 0 {
			columnCount = len(fields)
		}

		if len(fields) != columnCount {
			return nil, fmt.Errorf("importer: COPY row %d has %d fields, expected %d", line, len(fields), columnCount)
		}

		row := make([]any, columnCount)
		for i, field := range fields {
			value, err := decodeCopyField(field)
			if err != nil {
				return nil, fmt.Errorf("importer: decode COPY row %d column %d: %w", line, i+1, err)
			}
			row[i] = value
		}
		rows = append(rows, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("importer: read COPY data: %w", err)
	}

	return rows, nil
}

func decodeCopyField(field string) (any, error) {
	if field == "\\N" {
		return nil, nil
	}

	var builder strings.Builder

	for i := 0; i < len(field); i++ {
		ch := field[i]
		if ch != '\\' {
			builder.WriteByte(ch)
			continue
		}

		i++
		if i >= len(field) {
			return nil, errors.New("importer: COPY field ends with escape prefix")
		}

		next := field[i]
		switch next {
		case 'b':
			builder.WriteByte('\b')
		case 'f':
			builder.WriteByte('\f')
		case 'n':
			builder.WriteByte('\n')
		case 'r':
			builder.WriteByte('\r')
		case 't':
			builder.WriteByte('\t')
		case 'v':
			builder.WriteByte('\v')
		case '0':
			builder.WriteByte(0)
		case '.':
			builder.WriteByte('.')
		case '\\':
			builder.WriteByte('\\')
		default:
			if next >= '0' && next <= '7' {
				value := int(next - '0')
				consumed := 1
				for consumed < 3 && i+1 < len(field) {
					digit := field[i+1]
					if digit < '0' || digit > '7' {
						break
					}
					value = value*8 + int(digit-'0')
					i++
					consumed++
				}
				builder.WriteByte(byte(value))
				continue
			}
			builder.WriteByte(next)
		}
	}

	return builder.String(), nil
}

func sqlStateFromError(err error) (string, string) {
	if err == nil {
		return "", ""
	}

	var stateErr interface{ SQLState() string }
	if errors.As(err, &stateErr) {
		return stateErr.SQLState(), err.Error()
	}

	message := err.Error()
	upper := strings.ToUpper(message)
	marker := "(SQLSTATE "
	idx := strings.LastIndex(upper, marker)
	if idx != -1 {
		start := idx + len(marker)
		end := strings.Index(upper[start:], ")")
		if end != -1 {
			code := message[start : start+end]
			return code, message
		}
	}

	return "", message
}

func resolveCopyColumns(ctx context.Context, tx *sql.Tx, table string) ([]string, error) {
	schema, name := splitTableName(table)
	if name == "" {
		return nil, errors.New("importer: COPY statement missing target table name")
	}

	if schema == "" {
		schema = "public"
	}

	rows, err := tx.QueryContext(ctx, `SELECT column_name FROM information_schema.columns WHERE table_schema = $1 AND table_name = $2 ORDER BY ordinal_position`, schema, name)
	if err != nil {
		return nil, fmt.Errorf("importer: lookup columns for %s.%s: %w", schema, name, err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, fmt.Errorf("importer: scan column metadata: %w", err)
		}
		columns = append(columns, quoteIdentifier(col))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("importer: iterate column metadata: %w", err)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("importer: table %s.%s has no columns", schema, name)
	}

	return columns, nil
}

func splitTableName(identifier string) (string, string) {
	parts := splitQualifiedIdentifierParts(identifier)
	if len(parts) == 0 {
		return "", ""
	}

	if len(parts) == 1 {
		return "", unquoteIdentifierPart(parts[0])
	}

	schema := unquoteIdentifierPart(parts[0])
	name := unquoteIdentifierPart(parts[len(parts)-1])
	return schema, name
}

func splitQualifiedIdentifierParts(identifier string) []string {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil
	}

	data := []byte(identifier)
	var (
		parts   []string
		current strings.Builder
		inQuote bool
	)

	for i := 0; i < len(data); {
		r, size := utf8DecodeRune(data[i:])
		if size == 0 {
			break
		}

		if r == '"' {
			current.WriteRune(r)
			if inQuote && i+size < len(data) {
				next, nextSize := utf8DecodeRune(data[i+size:])
				if next == '"' {
					current.WriteRune(next)
					i += size + nextSize
					continue
				}
			}
			inQuote = !inQuote
			i += size
			continue
		}

		if r == '.' && !inQuote {
			part := strings.TrimSpace(current.String())
			if part != "" {
				parts = append(parts, part)
			}
			current.Reset()
			i += size
			continue
		}

		current.WriteRune(r)
		i += size
	}

	if current.Len() > 0 {
		part := strings.TrimSpace(current.String())
		if part != "" {
			parts = append(parts, part)
		}
	}

	return parts
}

func unquoteIdentifierPart(part string) string {
	part = strings.TrimSpace(part)
	if part == "" {
		return ""
	}

	if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") && len(part) >= 2 {
		inner := part[1 : len(part)-1]
		inner = strings.ReplaceAll(inner, `""`, `"`)
		return inner
	}

	return strings.ToLower(part)
}

func quoteIdentifier(identifier string) string {
	if identifier == "" {
		return identifier
	}

	needQuote := false
	for _, r := range identifier {
		if !(r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z') {
			needQuote = true
			break
		}
	}

	if !needQuote && identifier[0] >= 'a' && identifier[0] <= 'z' {
		return identifier
	}

	escaped := strings.ReplaceAll(identifier, `"`, `""`)
	return fmt.Sprintf("\"%s\"", escaped)
}

func statementTarget(sql string) (string, string) {
	data := []byte(sql)
	idx := skipIgnorableSections(data, 0)
	idx = skipLeadingCTESections(data, idx)
	idx = skipIgnorableSections(data, idx)
	trimmed := strings.TrimSpace(string(data[idx:]))
	trimmed = strings.TrimSpace(locateStatementStart(trimmed))
	if trimmed == "" {
		return "", ""
	}

	upper := strings.ToUpper(trimmed)

	switch {
	case strings.HasPrefix(upper, "COPY "):
		rest := trimmed[5:]
		token := firstIdentifier(rest, true)
		return token, "COPY"
	case strings.HasPrefix(upper, "INSERT INTO "):
		rest := trimmed[len("INSERT INTO "):]
		token := firstIdentifier(rest, true)
		return token, "INSERT"
	case strings.HasPrefix(upper, "UPDATE "):
		rest := trimmed[len("UPDATE "):]
		token := firstIdentifier(rest, false)
		return token, "UPDATE"
	case strings.HasPrefix(upper, "DELETE FROM "):
		rest := trimmed[len("DELETE FROM "):]
		token := firstIdentifier(rest, true)
		return token, "DELETE"
	case strings.HasPrefix(upper, "ALTER TABLE "):
		rest := trimmed[len("ALTER TABLE "):]
		token := firstIdentifier(rest, true)
		return token, "ALTER TABLE"
	case strings.HasPrefix(upper, "ALTER INDEX "):
		rest := trimmed[len("ALTER INDEX "):]
		token := firstIdentifier(rest, false)
		return token, "ALTER INDEX"
	case strings.HasPrefix(upper, "DROP TABLE "):
		rest := trimmed[len("DROP TABLE "):]
		token := firstIdentifier(rest, true)
		return token, "DROP TABLE"
	case strings.HasPrefix(upper, "DROP INDEX "):
		rest := trimmed[len("DROP INDEX "):]
		token := firstIdentifier(rest, false)
		return token, "DROP INDEX"
	case strings.HasPrefix(upper, "DROP SEQUENCE "):
		rest := trimmed[len("DROP SEQUENCE "):]
		token := firstIdentifier(rest, false)
		return token, "DROP SEQUENCE"
	case strings.HasPrefix(upper, "CREATE TABLE "):
		rest := trimmed[len("CREATE TABLE "):]
		token := firstIdentifier(rest, true)
		return token, "CREATE TABLE"
	case strings.HasPrefix(upper, "CREATE UNIQUE INDEX "):
		onIdx := strings.Index(upper, " ON ")
		if onIdx == -1 {
			return "", ""
		}
		rest := trimmed[onIdx+4:]
		token := firstIdentifier(rest, true)
		return token, "CREATE INDEX"
	case strings.HasPrefix(upper, "CREATE INDEX "):
		onIdx := strings.Index(upper, " ON ")
		if onIdx == -1 {
			return "", ""
		}
		rest := trimmed[onIdx+4:]
		token := firstIdentifier(rest, true)
		return token, "CREATE INDEX"
	case strings.HasPrefix(upper, "ALTER SEQUENCE "):
		rest := trimmed[len("ALTER SEQUENCE "):]
		token := firstIdentifier(rest, false)
		return token, "ALTER SEQUENCE"
	case strings.HasPrefix(upper, "SELECT ") && strings.Contains(upper, "SETVAL"):
		name := extractSetvalIdentifier(trimmed)
		if name != "" {
			return name, "SELECT"
		}
	}

	return "", ""
}

func locateStatementStart(input string) string {
	if input == "" {
		return input
	}

	data := []byte(input)
	keywords := []string{"COPY", "INSERT", "UPDATE", "DELETE", "ALTER", "DROP", "CREATE", "SELECT"}

	var (
		inSingleQuote  bool
		inDoubleQuote  bool
		inLineComment  bool
		inBlockComment bool
		dollarTag      string
	)

	for i := 0; i < len(data); {
		b := data[i]

		switch {
		case inLineComment:
			if b == '\n' {
				inLineComment = false
			}
			i++
			continue
		case inBlockComment:
			if b == '*' && i+1 < len(data) && data[i+1] == '/' {
				inBlockComment = false
				i += 2
				continue
			}
			i++
			continue
		case dollarTag != "":
			if b == '$' && hasDollarTag(data[i:], dollarTag) {
				i += len(dollarTag) + 2
				dollarTag = ""
				continue
			}
			i++
			continue
		case inSingleQuote:
			if b == '\\' && i+1 < len(data) {
				i += 2
				continue
			}
			if b == '\'' {
				inSingleQuote = false
			}
			i++
			continue
		case inDoubleQuote:
			if b == '"' {
				if i+1 < len(data) && data[i+1] == '"' {
					i += 2
					continue
				}
				inDoubleQuote = false
			}
			i++
			continue
		}

		if b == '-' && i+1 < len(data) && data[i+1] == '-' {
			inLineComment = true
			i += 2
			continue
		}

		if b == '/' && i+1 < len(data) && data[i+1] == '*' {
			inBlockComment = true
			i += 2
			continue
		}

		if b == '$' {
			if tag, ok := readDollarTag(data[i:]); ok {
				dollarTag = tag
				i += len(tag) + 2
				continue
			}
		}

		if b == '\'' {
			inSingleQuote = true
			i++
			continue
		}

		if b == '"' {
			inDoubleQuote = true
			i++
			continue
		}

		for _, kw := range keywords {
			if len(data)-i < len(kw) {
				continue
			}
			if !strings.EqualFold(string(data[i:i+len(kw)]), kw) {
				continue
			}

			if i > 0 {
				prev, _ := utf8DecodeLastRune(data[:i])
				if unicode.IsLetter(prev) || unicode.IsDigit(prev) || prev == '_' {
					goto noKeyword
				}
			}

			if i+len(kw) < len(data) {
				next, _ := utf8DecodeRune(data[i+len(kw):])
				if unicode.IsLetter(next) || unicode.IsDigit(next) || next == '_' {
					goto noKeyword
				}
			}

			return strings.TrimLeftFunc(string(data[i:]), unicode.IsSpace)
		}

	noKeyword:
		i++
	}

	return input
}

func skipLeadingCTESections(data []byte, idx int) int {
	original := idx
	idx = skipIgnorableSections(data, idx)
	if !hasPrefixFold(data[idx:], "WITH") {
		return original
	}

	cursor := idx + len("WITH")
	cursor = skipIgnorableSections(data, cursor)

	if next, ok := consumeKeyword(data, cursor, "RECURSIVE"); ok {
		cursor = next
		cursor = skipIgnorableSections(data, cursor)
	}

	for {
		cursor = skipIgnorableSections(data, cursor)
		next, ok := consumeIdentifier(data, cursor)
		if !ok {
			return original
		}
		cursor = next
		cursor = skipIgnorableSections(data, cursor)

		if cursor < len(data) && data[cursor] == '(' {
			end, ok := skipParentheticalSection(data, cursor)
			if !ok {
				return original
			}
			cursor = end + 1
			cursor = skipIgnorableSections(data, cursor)
		}

		if next, ok := consumeKeyword(data, cursor, "AS"); ok {
			cursor = next
		} else {
			return original
		}

		cursor = skipIgnorableSections(data, cursor)
		if next, ok := consumeKeyword(data, cursor, "NOT"); ok {
			cursor = next
			cursor = skipIgnorableSections(data, cursor)
			if next, ok := consumeKeyword(data, cursor, "MATERIALIZED"); ok {
				cursor = next
			} else {
				return original
			}
			cursor = skipIgnorableSections(data, cursor)
		} else if next, ok := consumeKeyword(data, cursor, "MATERIALIZED"); ok {
			cursor = next
			cursor = skipIgnorableSections(data, cursor)
		}

		if cursor >= len(data) || data[cursor] != '(' {
			return original
		}

		end, ok := skipParentheticalSection(data, cursor)
		if !ok {
			return original
		}
		cursor = end + 1
		cursor = skipIgnorableSections(data, cursor)

		if cursor < len(data) && data[cursor] == ',' {
			cursor++
			cursor = skipIgnorableSections(data, cursor)
			continue
		}

		break
	}

	cursor = skipIgnorableSections(data, cursor)
	return cursor
}

func hasPrefixFold(data []byte, prefix string) bool {
	if len(data) < len(prefix) {
		return false
	}
	return strings.EqualFold(string(data[:len(prefix)]), prefix)
}

func consumeKeyword(data []byte, idx int, keyword string) (int, bool) {
	idx = skipIgnorableSections(data, idx)
	if len(data[idx:]) < len(keyword) {
		return idx, false
	}

	if !strings.EqualFold(string(data[idx:idx+len(keyword)]), keyword) {
		return idx, false
	}

	next := idx + len(keyword)
	if next < len(data) {
		r, _ := utf8DecodeRune(data[next:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return idx, false
		}
	}

	return next, true
}

func consumeIdentifier(data []byte, idx int) (int, bool) {
	idx = skipIgnorableSections(data, idx)
	if idx >= len(data) {
		return idx, false
	}

	if data[idx] == '"' {
		idx++
		for idx < len(data) {
			if data[idx] == '"' {
				idx++
				if idx < len(data) && data[idx] == '"' {
					idx++
					continue
				}
				return idx, true
			}
			idx++
		}
		return idx, false
	}

	start := idx
	for idx < len(data) {
		r, size := utf8DecodeRune(data[idx:])
		if size == 0 {
			break
		}
		if r == '_' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			idx += size
			continue
		}
		break
	}

	if idx == start {
		return idx, false
	}

	return idx, true
}

func skipParentheticalSection(data []byte, idx int) (int, bool) {
	if idx >= len(data) || data[idx] != '(' {
		return idx, false
	}

	depth := 0
	i := idx
	inSingleQuote := false
	inDoubleQuote := false
	inLineComment := false
	inBlockComment := false
	dollarTag := ""

	for i < len(data) {
		b := data[i]

		switch {
		case inLineComment:
			if b == '\n' {
				inLineComment = false
			}
			i++
			continue
		case inBlockComment:
			if b == '*' && i+1 < len(data) && data[i+1] == '/' {
				inBlockComment = false
				i += 2
				continue
			}
			i++
			continue
		case dollarTag != "":
			if b == '$' && hasDollarTag(data[i:], dollarTag) {
				i += len(dollarTag) + 2
				dollarTag = ""
				continue
			}
			i++
			continue
		case inSingleQuote:
			if b == '\\' && i+1 < len(data) {
				i += 2
				continue
			}
			if b == '\'' {
				inSingleQuote = false
			}
			i++
			continue
		case inDoubleQuote:
			if b == '"' {
				if i+1 < len(data) && data[i+1] == '"' {
					i += 2
					continue
				}
				inDoubleQuote = false
			}
			i++
			continue
		}

		if b == '-' && i+1 < len(data) && data[i+1] == '-' {
			inLineComment = true
			i += 2
			continue
		}

		if b == '/' && i+1 < len(data) && data[i+1] == '*' {
			inBlockComment = true
			i += 2
			continue
		}

		if b == '$' {
			if tag, ok := readDollarTag(data[i:]); ok {
				dollarTag = tag
				i += len(tag) + 2
				continue
			}
		}

		if b == '\'' {
			inSingleQuote = true
			i++
			continue
		}

		if b == '"' {
			inDoubleQuote = true
			i++
			continue
		}

		if b == '(' {
			depth++
			i++
			continue
		}

		if b == ')' {
			if depth == 0 {
				return idx, false
			}
			depth--
			if depth == 0 {
				return i, true
			}
			i++
			continue
		}

		i++
	}

	return idx, false
}

func firstIdentifier(input string, allowOnly bool) string {
	s := strings.TrimLeftFunc(input, unicode.IsSpace)
	if allowOnly {
		if next, ok := trimLeadingKeyword(s, "ONLY"); ok {
			s = next
		}
	}

	prefixes := []string{"IF NOT EXISTS", "IF EXISTS"}
	for _, prefix := range prefixes {
		if next, ok := trimLeadingKeyword(s, prefix); ok {
			s = next
		}
	}

	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	if s == "" {
		return ""
	}

	var end int
	inQuotes := false
	for end < len(s) {
		c := s[end]
		if c == '"' {
			inQuotes = !inQuotes
			end++
			continue
		}
		if !inQuotes {
			if unicode.IsSpace(rune(c)) || c == '(' || c == ';' {
				break
			}
		}
		end++
	}

	return strings.TrimSpace(s[:end])
}

func trimLeadingKeyword(s, keyword string) (string, bool) {
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	if len(s) < len(keyword) {
		return s, false
	}

	candidate := s[:len(keyword)]
	if strings.EqualFold(candidate, keyword) {
		if len(s) == len(keyword) {
			return "", true
		}
		remainder := s[len(keyword):]
		if len(remainder) == 0 || unicode.IsSpace(rune(remainder[0])) {
			return remainder, true
		}
	}

	return s, false
}

var setvalRegex = regexp.MustCompile(`(?i)setval\s*\(\s*'([^']+)`)

func extractSetvalIdentifier(sql string) string {
	matches := setvalRegex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func shouldExcludeIdentifier(identifier string, skip map[string]struct{}) bool {
	normalized := normalizeQualifiedIdentifier(identifier)
	if normalized == "" {
		return false
	}

	if _, ok := skip[normalized]; ok {
		return true
	}

	parts := strings.Split(normalized, ".")
	last := parts[len(parts)-1]

	if _, ok := skip[last]; ok {
		return true
	}

	if strings.HasSuffix(last, "_id_seq") {
		base := strings.TrimSuffix(last, "_id_seq")
		if _, ok := skip[base]; ok {
			return true
		}
	}

	return false
}

func normalizeQualifiedIdentifier(identifier string) string {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return ""
	}

	parts := strings.Split(identifier, ".")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "\"")
		parts[i] = strings.ToLower(part)
	}

	return strings.Join(parts, ".")
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
