package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// var releaseTag string

func init() {
	rootCmd.AddCommand(prCmd)
	// prCmd.Flags().StringVar(&releaseTag, "releaseTag", "", "Tag for the release. If empty, curent date in YYYY.MM.DD will be used")
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
	compareBranch, err := gitRepo.CurrentBranch()
	if err != nil {
		fmt.Println(err)
	}

	pr, err := gc.CreatePR(baseBranch, compareBranch)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(pr)
}
