package main

import (
	"context"

	"github.com/ttrtcixy/users/internal/app"
)

func main() {
	ctx := context.Background()

	a := app.New(ctx)

	a.Run(context.Background())
}
