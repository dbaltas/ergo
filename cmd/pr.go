package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var compareBranch string
var title string
var description string

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.Flags().StringVar(&compareBranch, "compare", "", "The branch to compare with base branch.")
	prCmd.Flags().StringVar(&title, "title", "", "The title of the PR.")
	prCmd.Flags().StringVar(&description, "description", "", "The description of the PR.")
	prCmd.MarkFlagRequired("title")
}

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create a pull request",
	Long:  `Create a pull request from branch to baseBranch`,
	Run: func(cmd *cobra.Command, args []string) {
		createPR()
	},
}

func createPR() {
	if compareBranch != "" {
		var err error
		compareBranch, err = gitRepo.CurrentBranch()
		if err != nil {
			fmt.Println(err)
		}
	}

	pr, err := gc.CreatePR(baseBranch, compareBranch, title, description)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Created PR %s\n", *pr.HTMLURL)
}
