package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/greatbody/envman/internal/config"
	"github.com/greatbody/envman/internal/profile"
)

var (
	cfgMgr     *config.Manager
	profileMgr *profile.Manager
)

var rootCmd = &cobra.Command{
	Use:   "envman",
	Short: "Environment variable profile manager",
	Long:  "Manage and switch between different sets of environment variables.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cfgMgr = config.NewManager()
	profileMgr = profile.NewManager(cfgMgr.ProfilesDir())

	cobra.OnInitialize(func() {
		if err := cfgMgr.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing envman: %v\n", err)
			os.Exit(1)
		}
	})
}
