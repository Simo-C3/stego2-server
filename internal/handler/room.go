package handler

import (
	"net/http"

	"github.com/Simo-C3/stego2-server/internal/schema"
	"github.com/labstack/echo/v4"
)

type RoomHandler struct {
}

func NewRoomHandler() *RoomHandler {
	return &RoomHandler{}
}

func (h *RoomHandler) GetRooms(c echo.Context) error {
	rooms := []*schema.Room{
		{
			ID:         "1",
			Name:       "room1",
			HostName:   "host1",
			MinUserNum: 1,
			MaxUserNum: 4,
			UseCpu:     true,
		},
		{
			ID:         "2",
			Name:       "room2",
			HostName:   "host2",
			MinUserNum: 2,
			MaxUserNum: 4,
			UseCpu:     false,
		},
	}

	mockResponse := &schema.GetRoomsResponse{
		Rooms: rooms,
		Total: 2,
	}

	return c.JSON(http.StatusOK, mockResponse)
}

func (h *RoomHandler) CreateRoom(c echo.Context) error {

	req := new(schema.CreateRoomRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	mockResponse := &schema.CreateRoomResponse{
		RoomID: "1",
	}

	return c.JSON(http.StatusOK, mockResponse)
}

func (h *RoomHandler) Matching(c echo.Context) error {

	mockResponse := &schema.MatchingResponse{
		ID: "1",
	}

	return c.JSON(http.StatusOK, mockResponse)
}
