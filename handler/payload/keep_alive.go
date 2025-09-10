package payload

// KeepAliveResponse represents the response payload for the keep-alive endpoint.
type KeepAliveResponse struct {
	Message  string `json:"message"`
	DateTime string `json:"date_time"`
}
