package cli

import (
	"github.com/distraw/transaction-indexer/internal/config"
	"github.com/distraw/transaction-indexer/internal/core"

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
	jwtSecretFlag := serviceCMD.
		Flag("JTW_SECRET", "Secret to sign JSON Web Tokens with.").
		Envar("JWT_SECRET").
		Required().
		String()

	migrateCMD := app.Command("migrate", "migrate command")
	migrateUpCMD := migrateCMD.Command("up", "migrate db up")
	migrateDownCMD := migrateCMD.Command("down", "migrate db down")

	cmd, err := app.Parse(args[1:])
	if err != nil {
		log.WithError(err).Fatal("failed to parse arguments")
	}

	cfg := config.New(kv.MustFromEnv())

	switch cmd {
	case serviceCMD.FullCommand():
		err = core.RunServer(cfg, []byte(*jwtSecretFlag))
	case migrateUpCMD.FullCommand():
		err = MigrateUp(cfg)
	case migrateDownCMD.FullCommand():
		err = MigrateDown(cfg)
	default:
		log.WithError(err).Fatalf("unknown command %s", cmd)
	}

	if err != nil {
		log.WithError(err).Error("failed to exec cmd")
		return false
	}

	return true
}
