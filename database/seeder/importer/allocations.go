package importer

const (
	storageSQLDir         = "storage/sql"
	migrationsRelativeDir = "database/infra/migrations"
)

var excludedSeedTables = map[string]struct{}{}

type statement struct {
	sql      string
	copyData []byte
	isCopy   bool
}

type executeOptions struct {
	disableConstraints bool
	skipTables         map[string]struct{}
}
