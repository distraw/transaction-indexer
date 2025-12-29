package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/distraw/transaction-indexer/internal/config"

	"github.com/alecthomas/kingpin/v2"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
)

func Run(args []string) bool {
	log := logan.New()

	defer func() {
		if rvr := recover(); rvr != nil {
			log.WithRecover(rvr).Error("app panicked")
		}
	}()

	app := kingpin.New("transaction-indexer", "")
	runCMD := app.Command("run", "run command")

	serviceCMD := runCMD.Command("service", "run full service")

	migrateCMD := app.Command("migrate", "migrate command")
	migrateUpCMD := migrateCMD.Command("up", "migrate db up")
	migrateDownCMD := migrateCMD.Command("down", "migrate db down")

	cmd, err := app.Parse(args[1:])
	if err != nil {
		log.WithError(err).Fatal("failed to parse arguments")
	}

	_, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sigint := make(chan os.Signal, 1)

		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		cancel()
		os.Exit(0)
	}()

	cfg := config.New(kv.MustFromEnv())

	switch cmd {
	case serviceCMD.FullCommand():
		cfg.Log().Info("Hello, world!")
	case migrateUpCMD.FullCommand():
		err = MigrateUp(cfg)
	case migrateDownCMD.FullCommand():
		err = MigrateDown(cfg)
	default:
		log.WithError(err).Fatalf("unknown command %s", cmd)
	}

	if err != nil {
		log.WithError(err).Error("failed to exec cmd")
	}

	return true
}
