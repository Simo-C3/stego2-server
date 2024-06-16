package handler

import (
	"net/http"
	"strings"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/Simo-C3/stego2-server/internal/schema"
	"github.com/Simo-C3/stego2-server/pkg/middleware"
	"github.com/Simo-C3/stego2-server/pkg/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type RoomHandler struct {
	upgrader  *websocket.Upgrader
	wsHandler *WSHandler
	roomRepo  repository.RoomRepository
	otpRepo   repository.OTPRepository
	gameRepo  repository.GameRepository
}

func NewRoomHandler(wsHandler *WSHandler, roomRepo repository.RoomRepository, otpRepo repository.OTPRepository, gameRepo repository.GameRepository) *RoomHandler {
	return &RoomHandler{
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		wsHandler: wsHandler,
		roomRepo:  roomRepo,
		otpRepo:   otpRepo,
		gameRepo:  gameRepo,
	}
}

func convertToCreateRoomEntity(room *schema.CreateRoomRequest, uuid string, ownerID string) *model.Room {
	return &model.Room{
		ID:         uuid,
		OwnerID:    ownerID,
		Name:       room.Name,
		HostName:   room.HostName,
		MinUserNum: room.MinUserNum,
		MaxUserNum: room.MaxUserNum,
		UseCPU:     room.UseCPU,
		Status:     "pending",
	}
}

func convertToSchemaRoom(room *model.Room) *schema.Room {
	return &schema.Room{
		ID:         room.ID,
		OwnerID:    room.OwnerID,
		Name:       room.Name,
		HostName:   room.HostName,
		MinUserNum: room.MinUserNum,
		MaxUserNum: room.MaxUserNum,
		UseCPU:     room.UseCPU,
		Status:     room.Status,
	}
}

func (h *RoomHandler) GetRooms(c echo.Context) error {
	rooms, err := h.roomRepo.GetRooms(c.Request().Context())
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

	ownerID, err := middleware.GetUserID(c)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	createRoomRequest := convertToCreateRoomEntity(req, uuid, ownerID)

	roomID, err := h.roomRepo.CreateRoom(c.Request().Context(), createRoomRequest)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.JSON(http.StatusOK, schema.CreateRoomResponse{RoomID: roomID})
}

func (h *RoomHandler) Matching(c echo.Context) error {
	roomID, err := h.roomRepo.Matching(c.Request().Context())
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
	var req schema.JoinRoomQuery
	if err := c.Bind(&req); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	ctx := c.Request().Context()
	u, err := h.otpRepo.VerifyOTP(ctx, req.Otp)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid otp")
	}
	us := strings.SplitN(u, ";", 2)
	userID := us[0]
	displayName := us[1]

	game, err := h.gameRepo.GetGameByID(ctx, req.ID)
	if err != nil {
		room, err := h.roomRepo.GetRoomByID(ctx, req.ID)
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusNotFound, "room not found")
		}

		if room.Status != "pending" {
			return echo.NewHTTPError(http.StatusForbidden, "room is not pending")
		}

		game = model.NewGame(req.ID, model.GameStatusPending, room)
		if err := h.gameRepo.UpdateGame(ctx, game); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update game")
		}
	}

	_, userAlreadyExist := game.Users[userID]

	if !userAlreadyExist {
		user := model.NewUser(userID, displayName)
		if err := h.gameRepo.UpdateUser(ctx, user); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user")
		}

		if err := game.AddUser(user); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusForbidden, "failed to add user")
		}

		if err := h.gameRepo.UpdateGame(ctx, game); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update game")
		}
	}

	// Upgrade to websocket
	ws, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Errorf("%+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to upgrade to websocket")
	}
	defer ws.Close()

	h.wsHandler.Handle(ctx, ws, req.ID, userID, userAlreadyExist)

	return nil
}
