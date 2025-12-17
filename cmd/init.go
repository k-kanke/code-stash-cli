package cmd

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
)

var (
	initContextName  string
	initFolderID     string
	initCollectionID string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize codestash context for current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(initFolderID) == "" {
			return errors.New("folder id is required (use --folder)")
		}
		if strings.TrimSpace(initCollectionID) == "" {
			return errors.New("collection id is required (use --collection)")
		}

		st := requireState()
		ctxName := strings.TrimSpace(initContextName)
		if ctxName == "" {
			ctxName = "default"
		}

		st.SetContext(ctxName, initCollectionID, initFolderID)
		st.CurrentContext = ctxName
		if err := st.Save(); err != nil {
			return err
		}

		cmd.Printf("Initialized context %q with folder %s\n", ctxName, initFolderID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initCollectionID, "collection", "", "collection ID to bind")
	initCmd.Flags().StringVar(&initFolderID, "folder", "", "folder ID to bind")
	initCmd.Flags().StringVar(&initContextName, "context", "default", "context name")
}
