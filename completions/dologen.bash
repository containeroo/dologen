# bash completion for dologen
_dologen() {
  local cur prev opts
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  opts="--base64 --completion --help --password --password-file --server --username --version -b -f -h -p -s -u -v"

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
complete -F _dologen dologen
