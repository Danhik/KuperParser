package main

import (
	"context"
	"log"
	"time"

	"kuperparser/internal/config"
	"kuperparser/internal/logic"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка чтения config.yaml: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := logic.Run(ctx, cfg); err != nil {
		log.Fatalf("Ошибка выполнения: %v", err)
	}
}
