package handler

import (
	"net/http"

	"github.com/Simo-C3/stego2-server/internal/schema"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type RoomHandler struct {
	upgrader  *websocket.Upgrader
	wsHandler *WSHandler
}

func NewRoomHandler(wsHandler *WSHandler) *RoomHandler {
	return &RoomHandler{
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		wsHandler: wsHandler,
	}
}

// @Summary Get rooms
// @Description Get rooms
// @Tags rooms
// @Accept json
// @Produce json
// @Success 200 {object} schema.GetRoomsResponse
// @Failure 400 {object} schema.ErrResponse
// @Router /rooms [get]
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

// @Summary Create room
// @Description Create room
// @Tags rooms
// @Accept json
// @Produce json
// @Param request body schema.CreateRoomRequest true "Create room request"
// @Success 200 {object} schema.CreateRoomResponse
// @Failure 400 {object} schema.ErrResponse
// @Router /rooms/{room_id} [post]
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

// @Summary Room matching
// @Description Room matching
// @Tags rooms
// @Accept json
// @Produce json
// @Success 200 {object} schema.MatchingResponse
// @Failure 400 {object} schema.ErrResponse
// @Router /rooms/matching [get]
func (h *RoomHandler) Matching(c echo.Context) error {

	mockResponse := &schema.MatchingResponse{
		ID: "1",
	}

	return c.JSON(http.StatusOK, mockResponse)
}

func (h *RoomHandler) JoinRoom(c echo.Context) error {
	// Roomの存在を確認

	// ユーザーをRoomに追加

	// Upgrade to websocket
	ws, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Errorf("%+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to upgrade to websocket")
	}
	defer ws.Close()

	h.wsHandler.Handle(ws)

	return nil
}
