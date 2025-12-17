package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/k-kanke/code-stash-cli/internal/api"
	"github.com/k-kanke/code-stash-cli/internal/auth"
	"github.com/k-kanke/code-stash-cli/internal/config"
)

var notesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notes in the current context",
	RunE: func(cmd *cobra.Command, args []string) error {
		st := requireState()
		ctx, err := st.Current()
		if err != nil {
			return err
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
			return fmt.Errorf("not logged in; run `codestash login` first")
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
		if len(filtered) == 0 {
			cmd.Println("No notes found for this folder.")
			return nil
		}

		printNotesTable(cmd, filtered)
		return nil
	},
}

func init() {
	notesCmd.AddCommand(notesListCmd)
}

func filterNotesByFolder(notes []api.NoteSummary, folderID string) []api.NoteSummary {
	if strings.TrimSpace(folderID) == "" {
		return notes
	}
	result := make([]api.NoteSummary, 0, len(notes))
	for _, note := range notes {
		if note.FolderID == nil {
			continue
		}
		if *note.FolderID == folderID {
			result = append(result, note)
		}
	}
	return result
}

func printNotesTable(cmd *cobra.Command, notes []api.NoteSummary) {
	cmd.Printf("%-36s  %-30s  %-19s\n", "ID", "Title", "Updated")
	cmd.Println(strings.Repeat("-", 90))
	for _, note := range notes {
		updated := note.UpdatedAt.Format(time.RFC3339)
		title := note.Title
		if len(title) > 30 {
			title = title[:27] + "..."
		}
		cmd.Printf("%-36s  %-30s  %-19s\n", note.ID, title, updated)
	}
}
