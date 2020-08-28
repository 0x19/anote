package main

import (
	"os"

	"github.com/0x19/anote/cmd/anote/cmds"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func init() {
	var err error
	if logger, err = zap.NewProduction(); err != nil {
		panic(err)
	}
	cmds.SetLogger(logger)
}

func main() {
	defer logger.Sync()
	if err := cmds.GetRootCmd().Execute(); err != nil {
		logger.Error(
			"failed to execute root command", zap.Error(err),
		)
		os.Exit(1)
	}
}
