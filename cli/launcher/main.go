package main

import (
	_ "github.com/joho/godotenv/autoload"
)

import (
	"context"
	"github.com/HydroProtocol/hydro-sdk-backend/cli"
	"github.com/HydroProtocol/hydro-sdk-backend/launcher"
	"os"
)

func run() int {
	ctx, stop := context.WithCancel(context.Background())
	go cli.WaitExitSignal(stop)

	launcher.Run(ctx)
	return 0
}

func main() {
	os.Exit(run())
}
