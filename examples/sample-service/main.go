package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Postgres  string `json:"postgres"`
	Redis     string `json:"redis"`
	CheckedAt string `json:"checkedAt"`
}

func main() {
	port := getenv("PORT", "8080")

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/bookings", bookingsHandler)

	fmt.Printf("sample service listening on port %s\n", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("server failed: %v\n", err)
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	postgresStatus := checkPostgres(ctx)
	redisStatus := checkRedis(ctx)

	status := "ok"
	httpStatus := http.StatusOK

	if postgresStatus != "ok" || redisStatus != "ok" {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Postgres:  postgresStatus,
		Redis:     redisStatus,
		CheckedAt: time.Now().Format(time.RFC3339),
	}

	writeJSON(w, httpStatus, response)
}

func bookingsHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"bookings": []map[string]any{
			{
				"id":     1,
				"user":   "demo-user",
				"status": "confirmed",
			},
		},
	}

	writeJSON(w, http.StatusOK, response)
}

func checkPostgres(ctx context.Context) string {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return "missing DATABASE_URL"
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Sprintf("connect failed: %v", err)
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		return fmt.Sprintf("ping failed: %v", err)
	}

	return "ok"
}

func checkRedis(ctx context.Context) string {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return "missing REDIS_URL"
	}

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Sprintf("parse failed: %v", err)
	}

	client := redis.NewClient(options)
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Sprintf("ping failed: %v", err)
	}

	return "ok"
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		fmt.Printf("write response failed: %v\n", err)
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
