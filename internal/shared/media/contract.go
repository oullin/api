package media

type MultipartFormInterface interface {
	SetFile(file []byte)
	SetPayload(payload []byte)
	SetHeaderName(headerName string)
	GetFile() []byte
	GetPayload() []byte
	GetHeaderName() string
}
