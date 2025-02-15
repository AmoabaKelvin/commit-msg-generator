// AI COMMIT GENERATOR FOR FILES OR A SINGLE FILE

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// CommitResult stores the result for each file
type CommitResult struct {
	filename string
	message  string
	err      error
}

var helpMessages = map[string]string{
	"api-key": "Your API key (required). The default model is gpt-4o-mini and hence we need to provide an OPEN AI API KEY",
	"model":   "The model to use (e.g., davinci, gpt-3.5-turbo)",
	"r":       "Recursively process directories",
}

func confirmAction(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/n): ", prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
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
		log.Fatal("Error: Current directory is not a git repository")
	}

	_, err := getGitRoot()

	if err != nil {
		log.Fatal(err)
	}

	// todo: handle recursive mode later

	workingFile := flag.Arg(0)
	if workingFile == "" && !*recursive {
		log.Fatal("Error: No file provided and recursive flag is not set")
	}

	if workingFile != "" && !*recursive {
		// check if the file is Staged
		staged := isStaged(workingFile)

		if !staged {
			// log.Fatal("Error: File is not staged")

			// ask the user if they want to stage the file
			if confirmAction("File is not staged. Do you want to stage it?") {
				stageFile(workingFile)
			} else {
				log.Fatal("Error: File is not staged")
			}
		}
		diff, err := getDiffOfStagedFile(workingFile)
		if err != nil {
			log.Fatal("Error: Failed to get diff of staged file")
		}

		commitMessage, err := GenerateCommitMessage(diff)
		if err != nil {
			log.Fatal("Error: Failed to generate commit message")
		}

		// ask the user if they want to commit the message
		fmt.Println("Commit Message:", commitMessage)
		if confirmAction("Do you want to commit the message?") {
			if err := commit(commitMessage); err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("Sure, I'll not commit the message")
		}
	}

	if *recursive {
		files, err := getAllStagedFiles()
		if err != nil {
			log.Fatal("Error: Failed to get all staged files")
		}

		if len(files) == 0 {
			// ask if they want to stage all files
			if confirmAction("No staged files found. Do you want to stage all files?") {
				stageAllFiles()
			} else {
				log.Fatal("Error: No staged files found")
			}
		}

		if err := handleRecursiveMode(files); err != nil {
			log.Fatal(err)
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

func isStaged(filePath string) bool {
	output, err := exec.Command("git", "diff", "--cached", "--name-only", filePath).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), filePath)
}

func stageFile(filePath string) error {
	if err := exec.Command("git", "add", filePath).Run(); err != nil {
		return fmt.Errorf("failed to stage file %s: %w", filePath, err)
	}
	return nil
}

func stageAllFiles() error {
	if err := exec.Command("git", "add", ".").Run(); err != nil {
		return fmt.Errorf("failed to stage all files: %w", err)
	}
	return nil
}

// feature: commit the generated commit message
func commit(commitMessage string) error {
	if err := exec.Command("git", "commit", "-m", commitMessage).Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

func handleRecursiveMode(files []string) error {
	var wg sync.WaitGroup
	results := make([]CommitResult, len(files))

	// Process each file concurrently
	for i, file := range files {
		wg.Add(1)
		go func(index int, filepath string) {
			defer wg.Done()
			message, err := generateCommitMessageForFile(filepath)
			results[index] = CommitResult{
				filename: filepath,
				message:  message,
				err:      err,
			}
		}(i, file)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Display results with numbers
	for i, result := range results {
		if result.err != nil {
			fmt.Printf("%d. %s: Error - %v\n", i+1, result.filename, result.err)
			continue
		}
		fmt.Printf("%d. %s: %s\n", i+1, result.filename, result.message)
	}

	// Get user selection
	fmt.Println("\nEnter the numbers of commits you want to keep (comma-separated, e.g., '1,2,3'):")
	var input string
	fmt.Scanln(&input)

	// Process selected commits
	selectedIndices := parseSelection(input)
	for _, idx := range selectedIndices {
		if idx-1 < 0 || idx-1 >= len(results) {
			continue
		}
		result := results[idx-1]
		if result.err != nil {
			continue
		}
		if err := commit(result.message); err != nil {
			fmt.Printf("Failed to commit %s: %v\n", result.filename, err)
		} else {
			fmt.Printf("Successfully committed %s\n", result.filename)
		}
	}

	return nil
}

func parseSelection(input string) []int {
	parts := strings.Split(strings.TrimSpace(input), ",")
	var numbers []int
	for _, part := range parts {
		num, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		numbers = append(numbers, num)
	}
	return numbers
}
