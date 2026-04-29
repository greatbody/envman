package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/greatbody/envman/internal/shell"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Output shell init script for default profile",
	Long:  "Output shell commands to load the default profile. Add to .zshrc: eval $(envman init)",
	RunE: func(cmd *cobra.Command, args []string) error {
		defaultName, err := cfgMgr.GetDefaultProfile()
		if err != nil {
			return fmt.Errorf("getting default profile: %w", err)
		}

		p, err := profileMgr.Load(defaultName)
		if err != nil {
			return fmt.Errorf("loading default profile: %w", err)
		}

		output := shell.InitOutput(p.Vars)
		fmt.Fprintln(cmd.OutOrStdout(), output)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := profileMgr.List()
		if err != nil {
			return fmt.Errorf("listing profiles: %w", err)
		}

		defaultName, _ := cfgMgr.GetDefaultProfile()
		loaded := shell.GetLoadedProfiles()
		loadedMap := make(map[string]bool)
		if loaded != "" {
			for _, l := range splitComma(loaded) {
				loadedMap[l] = true
			}
		}

		for _, name := range names {
			marker := "  "
			if name == defaultName {
				marker = "* "
			}
			suffix := ""
			if loadedMap[name] {
				suffix = " [loaded]"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s\n", marker, name, suffix)
		}
		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show profile contents",
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

		masked, _ := cmd.Flags().GetBool("masked")

		for _, v := range p.Vars {
			if masked {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=****\n", v.Key)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", v.Key, v.Value)
			}
		}
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if profileMgr.Exists(name) {
			return fmt.Errorf("profile %s already exists", name)
		}

		if err := profileMgr.Create(name); err != nil {
			return fmt.Errorf("creating profile: %w", err)
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Profile %s created at %s\n", name, profileMgr.ProfilePath(name))
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Profile %s created. Opening in %s...\n", name, editor)
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm"},
	Short:   "Delete a profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !profileMgr.Exists(name) {
			return fmt.Errorf("profile %s does not exist", name)
		}

		defaultName, _ := cfgMgr.GetDefaultProfile()
		if name == defaultName {
			return fmt.Errorf("cannot delete the default profile")
		}

		if err := profileMgr.Delete(name); err != nil {
			return fmt.Errorf("deleting profile: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Profile %s deleted\n", name)
		return nil
	},
}

var copyCmd = &cobra.Command{
	Use:   "copy <source> <destination>",
	Short: "Copy a profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src, dst := args[0], args[1]
		if !profileMgr.Exists(src) {
			return fmt.Errorf("profile %s does not exist", src)
		}
		if profileMgr.Exists(dst) {
			return fmt.Errorf("profile %s already exists", dst)
		}

		if err := profileMgr.Copy(src, dst); err != nil {
			return fmt.Errorf("copying profile: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Profile %s copied to %s\n", src, dst)
		return nil
	},
}

var defaultCmd = &cobra.Command{
	Use:   "default <name>",
	Short: "Set the default profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !profileMgr.Exists(name) {
			return fmt.Errorf("profile %s does not exist", name)
		}

		if err := cfgMgr.SetDefaultProfile(name); err != nil {
			return fmt.Errorf("setting default profile: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Default profile set to %s\n", name)
		return nil
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff <profile-a> <profile-b>",
	Short: "Compare two profiles",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, b := args[0], args[1]
		if !profileMgr.Exists(a) {
			return fmt.Errorf("profile %s does not exist", a)
		}
		if !profileMgr.Exists(b) {
			return fmt.Errorf("profile %s does not exist", b)
		}

		added, removed, changed, err := profileMgr.Diff(a, b)
		if err != nil {
			return fmt.Errorf("comparing profiles: %w", err)
		}

		if len(added) == 0 && len(removed) == 0 && len(changed) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "Profiles are identical")
			return nil
		}

		if len(added) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Added in %s:\n", b)
			for _, k := range sortedKeys(added) {
				fmt.Fprintf(cmd.OutOrStdout(), "  + %s=%s\n", k, added[k])
			}
		}

		if len(removed) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Removed from %s:\n", b)
			for _, k := range sortedKeys(removed) {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s=%s\n", k, removed[k])
			}
		}

		if len(changed) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Changed:\n")
			for _, k := range sortedKeys(changed) {
				fmt.Fprintf(cmd.OutOrStdout(), "  ~ %s: %s -> %s\n", k, removed[k], changed[k])
			}
		}

		return nil
	},
}

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a profile in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !profileMgr.Exists(name) {
			return fmt.Errorf("profile %s does not exist", name)
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			return fmt.Errorf("$EDITOR is not set")
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Opening %s in %s...\n", name, editor)
		fmt.Fprintf(cmd.OutOrStdout(), "Path: %s\n", profileMgr.ProfilePath(name))
		return nil
	},
}

func init() {
	showCmd.Flags().Bool("masked", false, "Mask environment variable values")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(copyCmd)
	rootCmd.AddCommand(defaultCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(editCmd)
}

func splitComma(s string) []string {
	var result []string
	for _, part := range splitString(s, ',') {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s string, sep byte) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
