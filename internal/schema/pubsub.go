package schema

type PublishContent struct {
	RoomID  string `json:"roomID"`
	Payload any    `json:"payload"`
}
