package payload

type PingResponse struct {
	Message string `json:"message"`
	Date    string `json:"date"`
	Time    string `json:"time"`
}
