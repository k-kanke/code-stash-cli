package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/k-kanke/code-stash-cli/internal/api"
	"github.com/k-kanke/code-stash-cli/internal/auth"
	"github.com/k-kanke/code-stash-cli/internal/config"
	"github.com/k-kanke/code-stash-cli/internal/state"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage the active note scope",
}

var noteSwitchCmd = &cobra.Command{
	Use:   "switch <note-id>",
	Short: "Enter note scope for the given note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID := strings.TrimSpace(args[0])
		if noteID == "" {
			return errors.New("note id is required")
		}

		st := requireState()
		ctx, err := st.Current()
		if err != nil {
			return err
		}
		if st.Scope() == state.ScopeFolder {
			// ok, switching into note scope
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		token, err := auth.LoadToken(cfg.TokenPath)
		if err != nil {
			return err
		}
		if token == nil {
			return errors.New("not logged in; run `codestash login` first")
		}

		client, err := api.NewClient(cfg.APIBaseURL, cfg.ClientID, cfg.ClientSecret)
		if err != nil {
			return err
		}

		notes, err := client.ListNotes(cmd.Context(), token.AccessToken, ctx.Collection)
		if err != nil {
			return err
		}
		filtered := filterNotesByFolder(notes, ctx.Folder)

		var selected *api.NoteSummary
		for i := range filtered {
			n := filtered[i]
			if n.ID == noteID {
				selected = &n
				break
			}
		}
		if selected == nil {
			return fmt.Errorf("note %s not found in current folder", noteID)
		}

		if err := st.EnterNoteScope(selected.ID, selected.Title); err != nil {
			return err
		}
		if err := st.Save(); err != nil {
			return err
		}

		cmd.Printf("Switched to note %s (%s)\n", selected.Title, selected.ID)
		return nil
	},
}

var noteExitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Return to folder scope",
	RunE: func(cmd *cobra.Command, args []string) error {
		st := requireState()
		if st.Scope() != state.ScopeNote {
			cmd.Println("Already in folder scope.")
			return nil
		}
		st.EnterFolderScope()
		if err := st.Save(); err != nil {
			return err
		}
		cmd.Println("Exited note scope.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(noteCmd)
	noteCmd.AddCommand(noteSwitchCmd)
	noteCmd.AddCommand(noteExitCmd)
}
