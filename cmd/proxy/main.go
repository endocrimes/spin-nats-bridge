package main

import (
	"context"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/endocrimes/spin-nats-bridge/pkg/healthcheck"
	"github.com/endocrimes/spin-nats-bridge/pkg/system"
)

type cLI struct {
	O11yOTLP   string `name:"o11y-otlp" env:"O11Y_OTLP" default:"localhost:4317" help:"Address to send OTLP traces"`
	HealthAddr string `env:"HEALTH_ADDR" default:":10001" help:"The address for healthchecks to listen on"`
}

func main() {
	err := run()
	if err != nil {
		log.Printf("failed to run:\n\t%v\n", err)
		os.Exit(1)
	}
	log.Println("exiting 0")
}

func run() error {
	cli := cLI{}
	kong.Parse(&cli)

	ctx := context.Background()

	sys := system.New()
	defer sys.Cleanup(ctx)

	_, err := healthcheck.LoadIntoSystem(ctx, cli.HealthAddr, sys)
	if err != nil {
		return err
	}

	return sys.Run(ctx, 0)
}
