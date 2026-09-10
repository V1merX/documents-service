package main

import (
	"context"
	"log"

	"github.com/V1merX/documents-service/internal/app"
)

func main() {
	ctx := context.Background()

	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := a.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
