package handler

import (
	"net/http"

	"github.com/Simo-C3/stego2-server/internal/domain"
	"github.com/Simo-C3/stego2-server/internal/repository"
	"github.com/Simo-C3/stego2-server/internal/schema"
	"github.com/Simo-C3/stego2-server/pkg/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type RoomHandler struct {
	upgrader  *websocket.Upgrader
	wsHandler *WSHandler
	repo      *repository.RoomRepository
}

func NewRoomHandler(wsHandler *WSHandler, roomRepo *repository.RoomRepository) *RoomHandler {
	return &RoomHandler{
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		wsHandler: wsHandler,
		repo:      roomRepo,
	}
}

func convertToCreateRoomRequestDomainModel(room *schema.CreateRoomRequest) *domain.Room {
	return &domain.Room{
		ID:         uuid.GenerateUUIDv7(),
		Name:       room.Name,
		HostName:   room.HostName,
		MinUserNum: room.MinUserNum,
		MaxUserNum: room.MaxUserNum,
		UseCpu:     room.UseCpu,
	}
}

func convertToSchemaRoom(room *domain.Room) *schema.Room {
	return &schema.Room{
		ID:         room.ID,
		Name:       room.Name,
		HostName:   room.HostName,
		MinUserNum: room.MinUserNum,
		MaxUserNum: room.MaxUserNum,
		UseCpu:     room.UseCpu,
	}
}

func (h *RoomHandler) GetRooms(c echo.Context) error {
	rooms, err := h.repo.GetRooms(c.Request().Context())
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	res := make([]*schema.Room, 0, len(rooms))
	for _, room := range rooms {
		res = append(res, convertToSchemaRoom(room))
	}

	return c.JSON(http.StatusOK, res)
}

func (h *RoomHandler) CreateRoom(c echo.Context) error {
	req := new(schema.CreateRoomRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	createRoomRequest := convertToCreateRoomRequestDomainModel(req)

	roomRepo, err := h.repo.CreateRoom(c.Request().Context(), createRoomRequest)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.JSON(http.StatusOK, roomRepo)
}

func (h *RoomHandler) Matching(c echo.Context) error {
	roomID, err := h.repo.Matching(c.Request().Context())
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.JSON(http.StatusOK, roomID)
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
