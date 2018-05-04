package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dbaltas/ergo/github"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var releaseInterval string
var releaseOffset string

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.PersistentFlags().StringVar(&releaseOffset, "releaseOffset", "1m", "Duration to wait before the first release ('5m', '1h25m', '30s')")
	deployCmd.PersistentFlags().StringVar(&releaseInterval, "releaseInterval", "25m", "Duration to wait between releases. ('5m', '1h25m', '30s')")
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy base branch to target branches",
	Long:  `Deploy base branch to target branches`,
	Run: func(cmd *cobra.Command, args []string) {
		branchMap := viper.GetStringMapString("release.branch-map")

		getRepo()

		deployBranches(
			organizationName,
			viper.GetString("generic.remote"),
			repoForRelease,
			baseBranch,
			releaseBranches,
			releaseOffset,
			releaseInterval,
			directory,
			branchMap,
			viper.GetString("release.on-deploy.body-branch-suffix-find"),
			viper.GetString("release.on-deploy.body-branch-suffix-replace"),
			viper.GetString("github.access-token"),
		)
	},
}

func deployBranches(organization string, remote string, releaseRepo string, baseBranch string, branches []string, releaseOffset string, releaseInterval string, directory string, branchMap map[string]string, suffixFind string, suffixReplace string, githubAccessToken string) {
	blue := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)

	reference := baseBranch

	integrateGithubRelease := releaseRepo != ""

	if integrateGithubRelease {
		release, err := github.LastRelease(
			context.Background(),
			githubAccessToken,
			organization,
			releaseRepo,
		)
		if err != nil {
			fmt.Println(err)
		}
		reference = release.GetTagName()
		green.Printf("Deploying %s\n", release.GetHTMLURL())
	}

	blue.Printf("Release reference: %s\n", reference)
	green.Println("Deployment start times are estimates.")

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Branch", "Start Time")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	intervalDuration, err := time.ParseDuration(releaseInterval)
	if err != nil {
		fmt.Printf("error parsing interval %v", err)
		return
	}
	offsetDuration, err := time.ParseDuration(releaseOffset)
	if err != nil {
		fmt.Printf("error parsing offset %v", err)
		return
	}

	t := time.Now()
	t = t.Add(offsetDuration)
	firstRelease := t
	for _, branch := range branches {
		tbl.AddRow(branch, t.Format("15:04"))
		t = t.Add(intervalDuration)
	}

	tbl.Print()
	reader := bufio.NewReader(os.Stdin)
	yellow.Printf("Press 'ok' to continue with Deployment:")
	input, _ := reader.ReadString('\n')
	text := strings.Split(input, "\n")[0]
	if text != "ok" {
		fmt.Printf("No deployment\n")
		return
	}
	fmt.Println(text)

	if firstRelease.Before(time.Now()) {
		yellow.Println("\ndeployment stopped since first released time has passed. Please run again")
		return
	}

	d := firstRelease.Sub(time.Now())
	green.Printf("Deployment will start in %s\n", d.String())
	time.Sleep(d)

	for i, branch := range branches {
		if i != 0 {
			time.Sleep(intervalDuration)
			t = t.Add(intervalDuration)
		}
		green.Printf("%s Deploying %s\n", time.Now().Format("15:04:05"), branch)
		cmd := ""
		// if reference is a branch name, use origin
		if reference == baseBranch {
			cmd = fmt.Sprintf("cd %s && git push %s %s/%s:%s", directory, remote, remote, reference, branch)
		} else { // if reference is a tag don't prefix with origin
			cmd = fmt.Sprintf("cd %s && git push %s %s:%s", directory, remote, reference, branch)
		}
		green.Printf("%s Executing %s\n", time.Now().Format("15:04:05"), cmd)
		out, err := exec.Command("sh", "-c", cmd).Output()

		if err != nil {
			fmt.Printf("error executing command: %s %v\n", cmd, err)
			return
		}
		green.Printf("%s Triggered Successfully %s\n", time.Now().Format("15:04:05"), strings.TrimSpace(string(out)))

		branchText, ok := branchMap[branch]
		if integrateGithubRelease && ok && suffixFind != "" {
			t := time.Now()
			green.Printf("%s Updating release on github %s\n", time.Now().Format("15:04:05"), strings.TrimSpace(string(out)))
			release, err := github.LastRelease(
				context.Background(),
				githubAccessToken,
				organization,
				releaseRepo,
			)
			if err != nil {
				fmt.Println(err)
				return
			}

			findText := fmt.Sprintf("%s ![](https://img.shields.io/badge/released%s)", branchText, suffixFind)
			replaceText := fmt.Sprintf("%s ![](https://img.shields.io/badge/released-%d_%s_%d_%02d:%02d%s)", branchText, t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), suffixReplace)
			newBody := strings.Replace(*(release.Body), findText, replaceText, -1)
			fmt.Println(newBody)
			release.Body = &newBody
			release, err = github.EditRelease(
				context.Background(),
				githubAccessToken,
				organization,
				releaseRepo,
				release,
			)
			if err != nil {
				fmt.Println(err)
				return
			}

			green.Printf("%s Updated release on github %s\n", time.Now().Format("15:04:05"), strings.TrimSpace(string(out)))
		}
	}
}