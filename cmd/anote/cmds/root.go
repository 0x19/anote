package cmds

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	logger  *zap.Logger
	rootCmd = &cobra.Command{
		Use:   "anote",
		Short: "Anonutopia criptocurrency bot built on WAVES technology",
		Long: `
Anote is a Waves based cryptocurrency that can be mined, sent, received, exchanged or sold and used right from the start to purchase ads on our Telegram ad channel - AnonShout.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
)

// GetRootCmd returns the root cobra command
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// SetLogger sets the zapier logger to be cross used throughout the toolkit
func SetLogger(log *zap.Logger) {
	logger = log
}
