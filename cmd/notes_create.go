package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/k-kanke/code-stash-cli/internal/api"
	"github.com/k-kanke/code-stash-cli/internal/auth"
	"github.com/k-kanke/code-stash-cli/internal/config"
)

var (
	noteCreateTitle    string
	noteCreateLanguage string
	noteCreateTags     []string
	noteCreateFile     string
	noteCreateNoteFile string
)

var notesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new note in the current context",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFolderScope(); err != nil {
			return err
		}
		if strings.TrimSpace(noteCreateFile) == "" {
			return errors.New("--file is required")
		}
		if strings.TrimSpace(noteCreateTitle) == "" {
			return errors.New("--title is required")
		}

		st := requireState()
		ctx, err := st.Current()
		if err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		httpClient, err := api.NewClient(cfg.APIBaseURL, cfg.ClientID, cfg.ClientSecret)
		if err != nil {
			return err
		}

		absFile, err := filepath.Abs(noteCreateFile)
		if err != nil {
			return err
		}
		fileContent, err := os.ReadFile(absFile)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		var noteContent string
		if strings.TrimSpace(noteCreateNoteFile) != "" {
			noteAbs, err := filepath.Abs(noteCreateNoteFile)
			if err != nil {
				return err
			}
			body, err := os.ReadFile(noteAbs)
			if err != nil {
				return fmt.Errorf("read note body: %w", err)
			}
			noteContent = string(body)
		}

		token, err := auth.LoadToken(cfg.TokenPath)
		if err != nil {
			return err
		}
		if token == nil {
			return errors.New("not logged in; run `codestash login` first")
		}

		resp, err := httpClient.CreateNote(cmd.Context(), token.AccessToken, api.CreateNoteRequest{
			CollectionID: ctx.Collection,
			FolderID:     ctx.Folder,
			Title:        noteCreateTitle,
			Language:     noteCreateLanguage,
			Tags:         noteCreateTags,
			Code:         string(fileContent),
			Note:         noteContent,
		})
		if err != nil {
			return err
		}
		if resp == nil || resp.NoteID == "" {
			cmd.Println("Note created, but the server did not return an ID. Skipping local mapping.")
			return nil
		}

		st.SetFileMapping(ctx.Name, relativeToRoot(absFile), resp.NoteID)
		if err := st.Save(); err != nil {
			return err
		}

		cmd.Printf("Created note %q (ID: %s)\n", noteCreateTitle, resp.NoteID)
		return nil
	},
}

func init() {
	notesCmd.AddCommand(notesCreateCmd)

	notesCreateCmd.Flags().StringVar(&noteCreateTitle, "title", "", "note title")
	notesCreateCmd.Flags().StringVar(&noteCreateLanguage, "language", "", "code language")
	notesCreateCmd.Flags().StringSliceVar(&noteCreateTags, "tags", nil, "comma-separated tags")
	notesCreateCmd.Flags().StringVar(&noteCreateFile, "file", "", "path to code file")
	notesCreateCmd.Flags().StringVar(&noteCreateNoteFile, "note", "", "path to note/description file")
}
