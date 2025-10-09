package importer

const (
	storageSQLDir         = "storage/sql"
	migrationsRelativeDir = "database/infra/migrations"
)

var excludedSeedTables = map[string]struct{}{
	"api_keys":           {},
	"api_key_signatures": {},
}

type statement struct {
	sql      string
	copyData []byte
	isCopy   bool
}

type executeOptions struct {
	disableConstraints bool
	skipTables         map[string]struct{}
}
