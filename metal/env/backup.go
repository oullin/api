package env

type BackupEnvironment struct {
	Cron string `validate:"required,cron"`
	Dir  string `validate:"required"`
}
