package env

type NetEnvironment struct {
	HttpHost        string `validate:"required,lowercase,min=7"`
	HttpPort        string `validate:"required,numeric,oneof=8080"`
	PublicAllowedIP string `validate:"required_if=IsProduction true,omitempty,ip"`
	IsProduction    bool   `validate:"-"`
}

func (e NetEnvironment) GetHttpPort() string {
	return e.HttpPort
}

func (e NetEnvironment) GetHttpHost() string {
	return e.HttpHost
}

func (e NetEnvironment) GetHostURL() string {
	return e.HttpHost + ":" + e.HttpPort
}
