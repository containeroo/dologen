package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
)

const version = "1.2.4"

type registryAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
}

type dockerConfig struct {
	Auths map[string]registryAuth `json:"auths"`
}

func main() {
	os.Exit(run(os.Args[0], os.Args[1:], os.Stdout, os.Stderr))
}

func run(programName string, args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "completion" {
		if len(args) != 2 {
			fmt.Fprintf(stderr, "error: usage: %s completion <bash|zsh>\n", filepath.Base(programName))
			return 1
		}

		script, err := completionScript(filepath.Base(programName), args[1])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprint(stdout, script)
		return 0
	}

	fs := flag.NewFlagSet(programName, flag.ContinueOnError)
	fs.SetOutput(stderr)

	username := fs.StringP("username", "u", "", "username for docker registry")
	password := fs.StringP("password", "p", "", "password for docker registry")
	passwordFile := fs.StringP("password-file", "f", "", "password file for docker registry")
	server := fs.StringP("server", "s", "", "docker registry server")
	base64Output := fs.BoolP("base64", "b", false, "output result base64 encoded")
	completion := fs.String("completion", "", "print shell completion script: bash|zsh")
	printVersion := fs.BoolP("version", "v", false, "Print the current version and exit")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: %s [flags]\n", filepath.Base(programName))
		fmt.Fprintf(stderr, "       %s completion <bash|zsh>\n\n", filepath.Base(programName))
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *printVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}

	if *completion != "" {
		script, err := completionScript(filepath.Base(programName), *completion)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprint(stdout, script)
		return 0
	}

	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "error: unexpected arguments: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}

	if *username == "" {
		fmt.Fprintln(stderr, "error: username cannot be empty")
		fs.Usage()
		return 1
	}

	if *server == "" {
		fmt.Fprintln(stderr, "error: server cannot be empty")
		fs.Usage()
		return 1
	}

	passwordValue := *password
	if *passwordFile != "" {
		if *password != "" {
			fmt.Fprintln(stderr, "warning: both --password and --password-file were provided; using --password-file")
		}

		loadedPassword, warning, err := readPasswordFromFile(*passwordFile)
		if warning != "" {
			fmt.Fprintf(stderr, "warning: %s\n", warning)
		}
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		passwordValue = loadedPassword
	}

	if passwordValue == "" {
		fmt.Fprintln(stderr, "error: password cannot be empty")
		fs.Usage()
		return 1
	}

	result, err := buildDockerConfigJSON(*server, *username, passwordValue)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if *base64Output {
		result = base64.StdEncoding.EncodeToString([]byte(result))
	}

	fmt.Fprintln(stdout, result)
	return 0
}

func readPasswordFromFile(path string) (string, string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", "", fmt.Errorf("read password file error: %w", err)
	}

	if !fileInfo.Mode().IsRegular() {
		return "", "", fmt.Errorf("password file must be a regular file")
	}

	warning := ""
	if fileInfo.Mode().Perm()&0o077 != 0 {
		warning = fmt.Sprintf("password file %q is group/world accessible; consider chmod 600", path)
	}

	passwordBytes, err := os.ReadFile(path)
	if err != nil {
		return "", warning, fmt.Errorf("read password file error: %w", err)
	}

	password := strings.TrimRight(string(passwordBytes), "\r\n")
	if password == "" {
		return "", warning, fmt.Errorf("password file is empty")
	}

	return password, warning, nil
}

func buildDockerConfigJSON(server, username, password string) (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))

	config := dockerConfig{
		Auths: map[string]registryAuth{
			server: {
				Username: username,
				Password: password,
				Auth:     auth,
			},
		},
	}

	result, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("marshal docker config: %w", err)
	}

	return string(result), nil
}

func completionScript(binaryName, shell string) (string, error) {
	completerName := strings.NewReplacer("-", "_").Replace(binaryName)

	switch shell {
	case "bash":
		return fmt.Sprintf(`# bash completion for %s
_%s() {
  local cur prev opts
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  opts="completion --base64 --completion --help --password --password-file --server --username --version -b -f -h -p -s -u -v"

  if [[ ${COMP_CWORD} -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
    return 0
  fi

  if [[ "${COMP_WORDS[1]}" == "completion" ]]; then
    if [[ ${COMP_CWORD} -eq 2 ]]; then
      COMPREPLY=( $(compgen -W "bash zsh" -- "${cur}") )
      return 0
    fi
    return 0
  fi

  case "${prev}" in
    --completion)
      COMPREPLY=( $(compgen -W "bash zsh" -- "${cur}") )
      return 0
      ;;
    --password-file|-f)
      COMPREPLY=( $(compgen -f -- "${cur}") )
      return 0
      ;;
  esac

  COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
  return 0
}
complete -F _%s %s
`, binaryName, completerName, completerName, binaryName), nil
	case "zsh":
		return fmt.Sprintf(`#compdef %s

if (( CURRENT == 2 )) && [[ ${words[2]} != -* ]]; then
  _values 'command' completion
  return
fi

if [[ ${words[2]} == completion ]]; then
  if (( CURRENT == 3 )); then
    _values 'shell' bash zsh
  fi
  return
fi

_arguments -s \
  '(-u --username)'{-u,--username}'[username for docker registry]:username:' \
  '(-p --password)'{-p,--password}'[password for docker registry]:password:' \
  '(-f --password-file)'{-f,--password-file}'[password file for docker registry]:password file:_files' \
  '(-s --server)'{-s,--server}'[docker registry server]:server:' \
  '(-b --base64)'{-b,--base64}'[output result base64 encoded]' \
  '--completion[print shell completion script]:shell:(bash zsh)' \
  '(-v --version)'{-v,--version}'[Print the current version and exit]' \
  '(-h --help)'{-h,--help}'[help for %s]'
`, binaryName, binaryName), nil
	default:
		return "", fmt.Errorf("unsupported shell %q: expected bash or zsh", shell)
	}
}
