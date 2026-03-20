package env

type LogsEnvironment struct {
	Level      string `validate:"required,lowercase,oneof=debug info warn error"`
	Dir        string `validate:"required,lowercase"`
	DateFormat string `validate:"required,lowercase"`
}
