package cmd

import (
	"sort"

	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage codestash contexts",
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		st := requireState()
		names := make([]string, 0, len(st.Contexts))
		for name := range st.Contexts {
			names = append(names, name)
		}
		sort.Strings(names)

		current := st.CurrentContext
		for _, name := range names {
			ctx := st.Contexts[name]
			marker := " "
			if name == current {
				marker = "*"
			}
			cmd.Printf("%s %s (collection: %s, folder: %s)\n", marker, name, ctx.Collection, ctx.Folder)
		}
		if len(names) == 0 {
			cmd.Println("No contexts defined. Run `codestash init --folder <id>` to create one.")
		}
		return nil
	},
}

var contextSwitchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch active context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFolderScope(); err != nil {
			return err
		}
		st := requireState()
		name := args[0]
		if err := st.SwitchContext(name); err != nil {
			return err
		}
		if err := st.Save(); err != nil {
			return err
		}
		cmd.Printf("Switched to context %q\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextSwitchCmd)
}
