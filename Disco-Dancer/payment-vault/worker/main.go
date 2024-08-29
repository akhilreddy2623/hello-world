package main

import (
	"context"
	"time"

	"geico.visualstudio.com/Billing/plutus/logging"
)

var log = logging.GetLogger("payment-vault-worker")

func main() {
	ctx := context.Background()

	log.Info(ctx, "Starting payment vault worker role")

	keepAlive(ctx)
}

func keepAlive(ctx context.Context) {
	for {
		log.Info(ctx, "Keepalive is still running.")
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Minute):
		}
	}
}
