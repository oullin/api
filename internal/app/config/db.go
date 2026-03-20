package env

import "fmt"

type DBEnvironment struct {
	UserName     string `validate:"required,lowercase,min=10"`
	UserPassword string `validate:"required,min=10"`
	DatabaseName string `validate:"required,lowercase,min=5"`
	Port         int    `validate:"required,numeric,gt=0"`
	Host         string `validate:"required,lowercase,hostname"`
	DriverName   string `validate:"required,lowercase,oneof=postgres"`
	SSLMode      string `validate:"required,lowercase,oneof=require"`
	TimeZone     string `validate:"required"`
}

// GetDSN GORM DSN Connection String
// Example: "host=localhost user='user' password='pass' dbname='db' port=5432 sslmode=require TimeZone=Asia/Singapore"
func (e DBEnvironment) GetDSN() string {
	return fmt.Sprintf(
		"host=%s user='%s' password='%s' dbname='%s' port=%d sslmode=%s TimeZone=%s",
		e.Host,
		e.UserName,
		e.UserPassword,
		e.DatabaseName,
		e.Port,
		e.SSLMode,
		e.TimeZone,
	)
}
