package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/greatbody/envman/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Open interactive TUI for profile management",
	Long:  "Open a terminal UI to manage profiles. Use: eval $(envman tui)",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := profileMgr.List()
		if err != nil {
			return fmt.Errorf("listing profiles: %w", err)
		}

		defaultName, _ := cfgMgr.GetDefaultProfile()
		loaded := shellGetLoadedProfiles()

		model := tui.NewModel(names, defaultName, loaded, cfgMgr, profileMgr)
		p := tea.NewProgram(model, tea.WithOutput(os.Stderr))

		result, err := p.Run()
		if err != nil {
			return fmt.Errorf("running TUI: %w", err)
		}

		m, ok := result.(tui.Model)
		if !ok {
			return fmt.Errorf("unexpected model type")
		}

		if m.LoadOutput != "" {
			fmt.Fprintln(os.Stdout, m.LoadOutput)
		}

		return nil
	},
}

func shellGetLoadedProfiles() string {
	return os.Getenv("ENVMAN_LOADED_PROFILES")
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
