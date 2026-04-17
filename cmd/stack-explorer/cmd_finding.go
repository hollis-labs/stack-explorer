package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/chrispian/stack-explorer/internal/domain"
	"github.com/spf13/cobra"
)

var findingCmd = &cobra.Command{
	Use:   "finding",
	Short: "Manage research findings",
}

var findingAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a finding",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := &domain.Finding{}
		f.Title, _ = cmd.Flags().GetString("title")
		f.Category, _ = cmd.Flags().GetString("category")
		f.Severity, _ = cmd.Flags().GetString("severity")
		f.Description, _ = cmd.Flags().GetString("desc")
		f.Status = "open"

		repoID, _ := cmd.Flags().GetString("repo")
		if repoID != "" {
			f.RepoID = &repoID
		}

		if f.Title == "" {
			return fmt.Errorf("--title is required")
		}

		if err := store.CreateFinding(f); err != nil {
			return err
		}
		fmt.Printf("Finding #%d added: %s\n", f.ID, f.Title)
		return nil
	},
}

var findingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List findings",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoID, _ := cmd.Flags().GetString("repo")
		category, _ := cmd.Flags().GetString("category")
		status, _ := cmd.Flags().GetString("status")

		findings, err := store.ListFindings(repoID, category, status)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tREPO\tCATEGORY\tSEVERITY\tSTATUS\tTITLE\n")
		for _, f := range findings {
			repo := "-"
			if f.RepoID != nil {
				repo = *f.RepoID
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", f.ID, repo, f.Category, f.Severity, f.Status, f.Title)
		}
		w.Flush()
		return nil
	},
}

func init() {
	findingAddCmd.Flags().String("title", "", "finding title (required)")
	findingAddCmd.Flags().String("repo", "", "repo ID")
	findingAddCmd.Flags().String("category", "info", "gap|strength|opportunity|risk")
	findingAddCmd.Flags().String("severity", "info", "critical|high|medium|low|info")
	findingAddCmd.Flags().String("desc", "", "description")

	findingListCmd.Flags().String("repo", "", "filter by repo")
	findingListCmd.Flags().String("category", "", "filter by category")
	findingListCmd.Flags().String("status", "", "filter by status")

	findingCmd.AddCommand(findingAddCmd)
	findingCmd.AddCommand(findingListCmd)
}
