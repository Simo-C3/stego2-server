package schema


type (
	Room struct {
		Id string `json:"id"`
		Name string `json:"name"`
		HostName string `json:"hostName"`
		MinUserNum int `json:"minUserNum"`
		MaxUserNum int `json:"maxUserNum"`
		UseCpu bool `json:"useCpu"`
	}

	CreateRoomRequest struct {
		Name string `json:"name"`
		HostName string `json:"hostName"`
		MinUserNum int `json:"minUserNum"`
		MaxUserNum int `json:"maxUserNum"`
		UseCpu bool `json:"useCpu"`
	}

	CreateRoomResponse struct {
		RoomId string `json:"roomId"`
	}

	GetRoomsResponse struct {
		Rooms []Room `json:"rooms"`
		Total int `json:"total"`
	}

	MatchingResponse struct {
		Id string `json:"Id"`
	}
)
