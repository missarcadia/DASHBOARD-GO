package git

import (
	"os/exec"
	"strings"
)

// RunGitCommand executa um comando git e retorna sua saída ou um erro.
func RunGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetStatusInfo coleta várias informações de status do repositório Git.
func GetStatusInfo() map[string]string {
	status := make(map[string]string)

	changes, _ := RunGitCommand("status", "--porcelain")
	status["changes"] = changes

	branch, _ := RunGitCommand("branch", "--show-current")
	status["branch"] = branch

	lastLog, _ := RunGitCommand("log", "-1", "--oneline")
	status["last_log"] = lastLog

	return status
}
