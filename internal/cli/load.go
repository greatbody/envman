package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/greatbody/envman/internal/shell"
)

var loadCmd = &cobra.Command{
	Use:   "load <name>",
	Short: "Load a profile into the current shell session",
	Long:  "Output shell export commands. Use: eval $(envman load <name>)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !profileMgr.Exists(name) {
			return fmt.Errorf("profile %s does not exist", name)
		}

		p, err := profileMgr.Load(name)
		if err != nil {
			return fmt.Errorf("loading profile: %w", err)
		}

		loaded := shell.GetLoadedProfiles()
		output := shell.ExportWithTracking(p.Vars, name, loaded)
		fmt.Fprintln(cmd.OutOrStdout(), output)
		return nil
	},
}

var unloadCmd = &cobra.Command{
	Use:   "unload [name]",
	Short: "Unload a profile from the current shell session",
	Long:  "Output shell unset commands. Use: eval $(envman unload [name])",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		loaded := shell.GetLoadedProfiles()

		if all {
			output := shell.UnsetAll(loaded)
			fmt.Fprintln(cmd.OutOrStdout(), output)
			return nil
		}

		var name string
		if len(args) > 0 {
			name = args[0]
		} else {
			names := splitComma(loaded)
			if len(names) == 0 {
				return fmt.Errorf("no profiles loaded")
			}
			name = names[len(names)-1]
		}

		if !profileMgr.Exists(name) {
			return fmt.Errorf("profile %s does not exist", name)
		}

		p, err := profileMgr.Load(name)
		if err != nil {
			return fmt.Errorf("loading profile: %w", err)
		}

		output := shell.UnsetProfile(p.Vars, name, loaded)
		fmt.Fprintln(cmd.OutOrStdout(), output)
		return nil
	},
}

func init() {
	unloadCmd.Flags().Bool("all", false, "Unload all loaded profiles")

	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(unloadCmd)
}
