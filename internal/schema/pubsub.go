package schema

type PublishContent struct {
	RoomID       string   `json:"roomID"`
	Payload      any      `json:"payload"`
	IncludeUsers []string `json:"includeUsers"`
	ExcludeUsers []string `json:"excludeUsers"`
}
