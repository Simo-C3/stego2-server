package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

type Timer struct {
	ID     string
	Ticker *time.Ticker
	Done   chan bool
}

var (
	redisClient *redis.Client
	timers      = make(map[string]*Timer)
	mu          sync.Mutex
	ctx         = context.Background()
)

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/start", startGame)
	e.GET("/end", endGame)

	e.Logger.Fatal(e.Start(":50000"))
}

func startGame(c echo.Context) error {
	gameID := c.QueryParam("game")
	duration := c.QueryParam("duration")
	if gameID == "" {
		return c.String(http.StatusBadRequest, "game ID is required")
	}
	if duration == "" {
		return c.String(http.StatusBadRequest, "duration is required")
	}
	n, err := strconv.Atoi(duration)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid duration")
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := timers[gameID]; exists {
		return c.String(http.StatusBadRequest, "timer already exists for this game")
	}

	ticker := time.NewTicker(time.Duration(n) * time.Second)
	done := make(chan bool)
	timers[gameID] = &Timer{
		ID:     gameID,
		Ticker: ticker,
		Done:   done,
	}

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := redisClient.Publish(ctx, "game_notifications", fmt.Sprintf("Game %s: 30 seconds passed", gameID)).Err()
				if err != nil {
					log.Printf("Failed to publish message: %v", err)
				}
			}
		}
	}()

	return c.String(http.StatusOK, fmt.Sprintf("Timer started for game %s", gameID))
}

func endGame(c echo.Context) error {
	gameID := c.QueryParam("game")
	if gameID == "" {
		return c.String(http.StatusBadRequest, "game ID is required")
	}

	mu.Lock()
	defer mu.Unlock()

	timer, exists := timers[gameID]
	if !exists {
		return c.String(http.StatusBadRequest, "no timer found for this game")
	}

	timer.Ticker.Stop()
	timer.Done <- true
	delete(timers, gameID)

	return c.String(http.StatusOK, fmt.Sprintf("Timer stopped for game %s", gameID))
}
