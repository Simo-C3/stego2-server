package middleware

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
)

type AuthController interface {
	WithHeader(next echo.HandlerFunc) echo.HandlerFunc
	GetUser(c echo.Context) (*auth.UserRecord, error)
	GetUserByID(ctx context.Context, UserID string) (*auth.UserRecord, error)
}

type authController struct {
	client *auth.Client
}

func NewAuthController(ctx context.Context, cfg *config.FirebaseConfig) AuthController {
	cred := cfg.ServiceAccount
	var opt option.ClientOption
	if cred != "" {
		s, err := base64.StdEncoding.DecodeString(cred)
		if err != nil {
			log.Fatalf("firebase service account base64 decode error: %+v", err)
		}
		opt = option.WithCredentialsJSON(s)
	}

	app, err := firebase.NewApp(ctx, nil)
	if opt != nil {
		app, err = firebase.NewApp(ctx, nil, opt)
	}
	if err != nil {
		log.Fatalf("firebase app initialize error: %+v", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("firebase auth initialize error: %+v", err)
	}

	return &authController{
		client: client,
	}
}

const idTokenKey = "firebase-auth-idToken"

func (a *authController) WithHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := c.Request().Header.Get("Authorization")
		if auth == "" {
			c.Logger().Error("Authorization header required")
			return c.JSON(http.StatusUnauthorized, "Authorization header required")
		}

		if !strings.HasPrefix(auth, "Bearer ") {
			c.Logger().Error("Authorization header format must be Bearer {token}")
			return c.JSON(http.StatusUnauthorized, "Authorization header format must be Bearer {token}")
		}
		idToken := strings.TrimPrefix(auth, "Bearer ")

		ctx := c.Request().Context()
		token, err := a.client.VerifyIDToken(ctx, idToken)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusUnauthorized, err.Error())
		}
		c.Set(idTokenKey, token)

		return next(c)
	}
}

func GetUserID(c echo.Context) (string, error) {
	if token, ok := c.Get(idTokenKey).(*auth.Token); ok {
		return token.UID, nil
	}
	return "", echo.NewHTTPError(http.StatusInternalServerError, "uid not found")
}

func (a *authController) GetUser(c echo.Context) (*auth.UserRecord, error) {
	uid, err := GetUserID(c)
	if err != nil {
		return nil, err
	}

	ctx := c.Request().Context()
	user, err := a.GetUserByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *authController) GetUserByID(ctx context.Context, UserID string) (*auth.UserRecord, error) {
	user, err := a.client.GetUser(ctx, UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
