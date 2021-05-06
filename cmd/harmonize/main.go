package main

import (
	"log"
	"os"

	"github.com/concourse/governance"
	"github.com/concourse/governance/cmd/harmonize/delta"
	"go.uber.org/zap"
)

// Concourse Discord server ID
const guildID = "219899946617274369"

func main() {
	logger, err := zap.NewDevelopment(zap.IncreaseLevel(zap.InfoLevel))
	if err != nil {
		log.Fatalln("zap:", err)
	}

	defer logger.Sync()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logger.Fatal("no $DISCORD_TOKEN provided")
	}

	discord, err := delta.NewDiscord(guildID, token)
	if err != nil {
		logger.Fatal("failed to initialize discord", zap.Error(err))
	}

	if os.Getenv("DISCORD_DRY_RUN") != "" {
		logger.Info("performing dry run")

		discord = dryRunDiscord{discord}
	}

	config, err := governance.LoadConfig(os.DirFS("."))
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	diff, err := delta.Diff(config, discord)
	if err != nil {
		logger.Fatal("failed to compute diff", zap.Error(err))
	}

	if len(diff) == 0 {
		logger.Info("nothing to do")
		return
	}

	for _, delta := range diff {
		err := delta.Apply(logger, discord)
		if err != nil {
			logger.Sugar().Fatalf("failed to apply %T: %s", delta, err)
		}
	}
}

type dryRunDiscord struct {
	delta.Discord
}

func (discord dryRunDiscord) CreateRole(delta.DeltaRoleCreate) error          { return nil }
func (discord dryRunDiscord) EditRole(delta.DeltaRoleEdit) error              { return nil }
func (discord dryRunDiscord) SetRolePositions(delta.DeltaRolePositions) error { return nil }
func (discord dryRunDiscord) AddUserRole(delta.DeltaUserAddRole) error        { return nil }
func (discord dryRunDiscord) RemoveUserRole(delta.DeltaUserRemoveRole) error  { return nil }
