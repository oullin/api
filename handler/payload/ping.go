package payload

type PingResponse struct {
	Message  string `json:"message"`
	DateTime string `json:"date_time"`
}
