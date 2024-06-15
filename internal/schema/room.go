package schema

type (
	Room struct {
		ID         string `json:"id"`
		OwnerID    string `json:"ownerId"`
		Name       string `json:"name"`
		HostName   string `json:"hostName"`
		MinUserNum int    `json:"minUserNum"`
		MaxUserNum int    `json:"maxUserNum"`
		UseCPU     bool   `json:"useCpu"`
		Status     string `json:"status"`
	}

	CreateRoomRequest struct {
		Name       string `json:"name"`
		HostName   string `json:"hostName"`
		MinUserNum int    `json:"minUserNum"`
		MaxUserNum int    `json:"maxUserNum"`
		UseCPU     bool   `json:"useCpu"`
	}

	CreateRoomResponse struct {
		RoomID string `json:"roomId"`
	}

	GetRoomsResponse struct {
		Rooms []*Room `json:"rooms"`
		Total int     `json:"total"`
	}

	MatchingResponse struct {
		ID string `json:"id"`
	}
)
