package handler

import (
	"net/http"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/Simo-C3/stego2-server/internal/schema"
	"github.com/Simo-C3/stego2-server/pkg/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type RoomHandler struct {
	upgrader  *websocket.Upgrader
	wsHandler *WSHandler
	repo      repository.RoomRepository
}

func NewRoomHandler(wsHandler *WSHandler, roomRepo repository.RoomRepository) *RoomHandler {
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

func convertToCreateRoomEntity(room *schema.CreateRoomRequest, uuid string) *model.Room {
	return &model.Room{
		ID:         uuid,
		Name:       room.Name,
		HostName:   room.HostName,
		MinUserNum: room.MinUserNum,
		MaxUserNum: room.MaxUserNum,
		UseCPU:     room.UseCPU,
	}
}

func convertToSchemaRoom(room *model.Room) *schema.Room {
	return &schema.Room{
		ID:         room.ID,
		Name:       room.Name,
		HostName:   room.HostName,
		MinUserNum: room.MinUserNum,
		MaxUserNum: room.MaxUserNum,
		UseCPU:     room.UseCPU,
		Status:     room.Status,
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

	uuid, err := uuid.GenerateUUIDv7()
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	createRoomRequest := convertToCreateRoomEntity(req, uuid)

	roomID, err := h.repo.CreateRoom(c.Request().Context(), createRoomRequest)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.JSON(http.StatusOK, schema.CreateRoomResponse{RoomID: roomID})
}

func (h *RoomHandler) Matching(c echo.Context) error {
	roomID, err := h.repo.Matching(c.Request().Context())
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	if roomID == "" {
		return c.JSON(http.StatusNotFound, "no room found")
	}

	return c.JSON(http.StatusOK, schema.MatchingResponse{ID: roomID})
}

func (h *RoomHandler) JoinRoom(c echo.Context) error {
	// Roomの存在を確認
	roomID := c.Param("id")

	ctx := c.Request().Context()
	_, err := h.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// ユーザーをRoomに追加

	// Upgrade to websocket
	ws, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Errorf("%+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to upgrade to websocket")
	}
	defer ws.Close()

	h.wsHandler.Handle(ctx, ws, roomID, "dummyUserID")

	return nil
}
