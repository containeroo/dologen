#compdef dologen

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
  '(-h --help)'{-h,--help}'[help for dologen]'
