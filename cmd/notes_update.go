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
	noteUpdateFile     string
	noteUpdateTitle    string
	noteUpdateLang     string
	noteUpdateTags     []string
	noteUpdateNoteFile string
)

var notesUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing note based on a local file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireNoteScope(); err != nil {
			return err
		}
		if strings.TrimSpace(noteUpdateFile) == "" {
			return errors.New("--file is required")
		}

		st := requireState()
		noteID, noteTitle, err := st.CurrentNote()
		if err != nil {
			return err
		}

		absFile, err := filepath.Abs(noteUpdateFile)
		if err != nil {
			return err
		}

		fileContent, err := os.ReadFile(absFile)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		var noteContent *string
		if strings.TrimSpace(noteUpdateNoteFile) != "" {
			noteAbs, err := filepath.Abs(noteUpdateNoteFile)
			if err != nil {
				return err
			}
			body, err := os.ReadFile(noteAbs)
			if err != nil {
				return fmt.Errorf("read note body: %w", err)
			}
			content := string(body)
			noteContent = &content
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

		req := api.UpdateNoteRequest{}

		codeStr := string(fileContent)
		req.Code = &codeStr

		if cmd.Flags().Changed("title") {
			title := strings.TrimSpace(noteUpdateTitle)
			if title == "" {
				return errors.New("--title cannot be empty when provided")
			}
			req.Title = &title
		}

		if cmd.Flags().Changed("language") {
			lang := strings.TrimSpace(noteUpdateLang)
			if lang == "" {
				return errors.New("--language cannot be empty when provided")
			}
			req.Language = &lang
		}

		if cmd.Flags().Changed("tags") {
			req.Tags = noteUpdateTags
		}

		if noteContent != nil {
			req.Note = noteContent
		}

		if err := client.UpdateNote(cmd.Context(), token.AccessToken, noteID, req); err != nil {
			return err
		}

		target := noteID
		if noteTitle != "" {
			target = fmt.Sprintf("%s (%s)", noteTitle, noteID)
		}
		cmd.Printf("Updated note %s\n", target)
		return nil
	},
}

func init() {
	notesCmd.AddCommand(notesUpdateCmd)

	notesUpdateCmd.Flags().StringVar(&noteUpdateFile, "file", "", "path to code file")
	notesUpdateCmd.Flags().StringVar(&noteUpdateTitle, "title", "", "new title")
	notesUpdateCmd.Flags().StringVar(&noteUpdateLang, "language", "", "code language")
	notesUpdateCmd.Flags().StringSliceVar(&noteUpdateTags, "tags", nil, "comma-separated tags")
	notesUpdateCmd.Flags().StringVar(&noteUpdateNoteFile, "note", "", "path to note/description file")
}
