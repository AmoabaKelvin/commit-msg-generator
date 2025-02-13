// AI COMMIT GENERATOR FOR FILES OR A SINGLE FILE

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var helpMessages = map[string]string{
	"api-key": "Your API key (required). The default model is gpt-4o-mini and hence we need to provide an OPEN AI API KEY",
	"model":   "The model to use (e.g., davinci, gpt-3.5-turbo)",
	"r":       "Recursively process directories",
}

func main() {
	apiKey := flag.String("api-key", "", helpMessages["api-key"])
	// model := flag.String("model", "gpt-4o-mini", helpMessages["model"])
	recursive := flag.Bool("r", false, helpMessages["r"])

	flag.Parse()

	if *apiKey == "" && os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Println("Error: --api-key is required")
		flag.Usage()
		os.Exit(1)
	}

	if !isGitRepo() {
		fmt.Println("Error: Current directory is not a git repository")
		os.Exit(1)
	}

	_, err := getGitRoot()

	if err != nil {
		fmt.Println("Error: Failed to get git root")
		os.Exit(1)
	}

	// todo: handle recursive mode later

	stagedFile := flag.Arg(0)
	if stagedFile == "" && !*recursive {
		fmt.Println("Error: No file provided and recursive flag is not set")
		os.Exit(1)
	}

	if stagedFile != "" && !*recursive {
		diff, err := getDiffOfStagedFile(stagedFile)
		if err != nil {
			fmt.Println("Error: Failed to get diff of staged file")
			os.Exit(1)
		}

		commitMessage, err := GenerateCommitMessage(diff)
		if err != nil {
			fmt.Println("Error: Failed to generate commit message")
			os.Exit(1)
		}

		fmt.Println("Commit Message:", commitMessage)
	}

	if *recursive {
		files, err := getAllStagedFiles()
		if err != nil {
			fmt.Println("Error: Failed to get all staged files")
			os.Exit(1)
		}

		if len(files) == 0 {
			log.Fatal("No staged files found")
		}

		for _, file := range files {
			commitMessage, err := generateCommitMessageForFile(file)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Commit Message: %s\n", commitMessage)
		}
	}
}

func isGitRepo() bool {
	return exec.Command("git", "rev-parse", "--is-inside-work-tree").Run() == nil
}

func getGitRoot() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getDiffOfStagedFile(filePath string) (string, error) {
	output, err := exec.Command("git", "diff", "--cached", filePath).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getAllStagedFiles() ([]string, error) {
	output, err := exec.Command("git", "diff", "--cached", "--name-only").Output()
	if err != nil {
		return nil, err
	}
	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, f := range files {
		if f != "" {
			result = append(result, f)
		}
	}
	return result, nil
}

func generateCommitMessageForFile(filePath string) (string, error) {
	diff, err := getDiffOfStagedFile(filePath)
	if err != nil {
		return "", err
	}

	commitMessage, err := GenerateCommitMessage(diff)
	if err != nil {
		return "", err
	}
	return commitMessage, nil
}
