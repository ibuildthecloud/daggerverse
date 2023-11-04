package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/ibuildthecloud/dagamole/pkg/bootstrap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := bootstrap.Main(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
