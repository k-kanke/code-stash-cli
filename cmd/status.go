package cmd

import (
	"github.com/spf13/cobra"

	"github.com/k-kanke/code-stash-cli/internal/state"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current codestash context and scope",
	RunE: func(cmd *cobra.Command, args []string) error {
		st := requireState()
		ctx, err := st.Current()
		if err != nil {
			return err
		}
		scope := st.Scope()

		cmd.Printf("Context: %s (collection: %s, folder: %s)\n", ctx.Name, ctx.Collection, ctx.Folder)
		cmd.Printf("Scope: %s\n", scope)

		if scope == state.ScopeNote {
			noteID, noteTitle, err := st.CurrentNote()
			if err != nil {
				return err
			}
			if noteTitle != "" {
				cmd.Printf("Note: %s (%s)\n", noteTitle, noteID)
			} else {
				cmd.Printf("Note: %s\n", noteID)
			}
			cmd.Println("Available commands: notes update, note exit, notes list, status")
		} else {
			cmd.Println("Note: <none>")
			cmd.Println("Available commands: notes create, notes list, note switch, context switch, status")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
