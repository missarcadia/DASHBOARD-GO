package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// RunGitCommand executa um comando git e retorna sua saída ou um erro.
func RunGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Retorna a saída mesmo em caso de erro, pois o Git geralmente imprime a mensagem de erro nela
		return string(output), err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetDetailedStatus executa 'git status --porcelain' e formata a saída.
func GetDetailedStatus() (string, error) {
	output, err := RunGitCommand("status", "--porcelain")
	if err != nil {
		return "", err
	}
	if output == "" {
		return "✅  Nenhuma alteração no diretório de trabalho.", nil
	}

	var builder strings.Builder
	builder.WriteString("Alterações não preparadas para commit:\n")
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		file := line[3:]

		switch status {
		case " M":
			builder.WriteString(fmt.Sprintf("  📝 Modificado: %s\n", file))
		case " A":
			builder.WriteString(fmt.Sprintf("  ➕ Adicionado (Staged): %s\n", file))
		case "??":
			builder.WriteString(fmt.Sprintf("  ❓ Não rastreado: %s\n", file))
		default:
			builder.WriteString(fmt.Sprintf("  %s %s\n", status, file))
		}
	}

	return builder.String(), nil
}

// GetGitUserInfo busca o nome, email e @ do usuário do GitHub a partir da config e remote.
func GetGitUserInfo() (string, error) {
	name, _ := RunGitCommand("config", "user.name")
	email, _ := RunGitCommand("config", "user.email")
	remoteURL, err := RunGitCommand("remote", "get-url", "origin")

	if err != nil {
		return "Repositório remoto 'origin' não configurado.", nil
	}

	// Expressão regular para extrair o usuário da URL do GitHub
	re := regexp.MustCompile(`github\.com[:/]([^/]+)`)
	matches := re.FindStringSubmatch(remoteURL)

	githubUser := "Não encontrado"
	if len(matches) > 1 {
		githubUser = matches[1]
	}

	var builder strings.Builder
	builder.WriteString("👤 Informações do Usuário Git\n")
	builder.WriteString("---------------------------\n")
	builder.WriteString(fmt.Sprintf("Nome:  %s\n", name))
	builder.WriteString(fmt.Sprintf("Email: %s\n", email))
	builder.WriteString(fmt.Sprintf("Conta: @%s\n", githubUser))

	return builder.String(), nil
}
