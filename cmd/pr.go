package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var compareBranch string
var title string
var description string
var number int

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.Flags().StringVar(&compareBranch, "compare", "", "The branch to compare with base branch. Defaults to current local branch.")
	prCmd.Flags().StringVar(&title, "title", "", "The title of the PR.")
	prCmd.Flags().StringVar(&description, "description", "", "The description of the PR.")
	// prCmd.MarkFlagRequired("title")
}

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create a pull request [github]",
	Long:  `Create a pull request on github from compare branch to base branch`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if len(args) > 0 {
			number, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			return getPR(number)
		}

		return createPR()
	},
}

func createPR() error {
	var err error
	yellow := color.New(color.FgYellow)

	if compareBranch == "" {
		compareBranch, err = r.CurrentBranch()
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Printf(`Create a PR
	base:%s
	compare:%s
	title:%s
	description:%s
`, baseBranch, compareBranch, title, description)

	yellow.Printf("\nPress 'ok' to continue:")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	text := strings.Split(input, "\n")[0]
	if text != "ok" {
		fmt.Printf("No PR\n")
		return nil
	}

	pr, err := gc.CreatePR(baseBranch, compareBranch, title, description)

	if err != nil {
		return err
	}

	fmt.Printf("Created PR %s\n", *pr.HTMLURL)

	return nil
}

func getPR(number int) error {
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	prp, err := gc.GetPR(number)
	pr := *prp

	if err != nil {
		return err
	}

	yellow.Printf("into:%s from:%s\n", pr.Base.GetLabel(), pr.Head.GetLabel())
	green.Printf("#%d: %s\n", pr.GetNumber(), pr.GetTitle())
	if pr.GetBody() != "" {
		yellow.Println(pr.GetBody())
	}

	a := green.Sprintf("%d", pr.GetAdditions())
	d := red.Sprintf("%d", pr.GetDeletions())
	c := yellow.Sprintf("%d", pr.GetChangedFiles())
	// var userNames []string
	// for _, usr := range pr. {
	// 	userNames = append(userNames, usr.GetLogin())
	// }
	// fmt.Printf("Reviewers: %s\n", strings.Join(userNames, ", "))
	fmt.Println(pr.GetReviewCommentsURL())
	fmt.Println(pr.GetCommits())
	fmt.Println(pr.GetCreatedAt())
	fmt.Println(pr.GetUpdatedAt())
	fmt.Println(pr.GetReviewCommentsURL())
	fmt.Println(pr.GetReviewCommentsURL())
	fmt.Printf("%s files changed, %s additions, %s deletions, %s comments\n", c, a, d, pr.GetComments())
	return nil
}
