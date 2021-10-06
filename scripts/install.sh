#!/usr/bin/env bash

PROGRAM_NAME="hz-cli"
HZCLI_HOME="$HOME/.local/share/hz-cli"

ghExtractTag() {
  tagUrl=$(curl "https://github.com/$1/releases/latest" -s -L -I -o /dev/null -w '%{url_effective}')
  printf "%s\n" "${tagUrl##*v}"
}

bin_id=""
machine=$(uname -m)

case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
    "linux")
        case "$machine" in
            "arm64"*) bin_id='Linux_arm64' ;;
            *"x86_64") bin_id='Linux_x86_64' ;;
        esac
    ;;
    "darwin")
        case "$machine" in
            *"arm64") bin_id='Darwin_arm64' ;;
            *"x86_64") bin_id='Darwin_x86_64' ;;
        esac
    ;;
esac

tag=$(ghExtractTag hazelcast/hazelcast-commandline-client)
releaseUrl=$(printf "https://github.com/hazelcast/hazelcast-commandline-client/releases/download/v%s/hz-cli_%s_%s" "$tag" "$tag" "$bin_id")

curl -L --silent "$releaseUrl" --output "$HOME/$PROGRAM_NAME"
chmod +x $HOME/$PROGRAM_NAME

mkdir -p $HOME/.local/bin
mv $HOME/hz-cli $HOME/.local/bin
echo "Hazelcast Commandline Client (CLC) is downloaded to \$HOME/.local/bin/$PROGRAM_NAME"
echo

read -rd '' addToPathDirectivesZSH << EOF
* Add \$HOME/.local/bin to PATH to access hz-cli from any directory
  To add it to your path, simply execute:
  echo "export PATH=\$HOME/.local/bin:\$PATH" >> \$HOME/.zshrc

EOF

read -rd '' addToPathDirectivesBASH << EOF
* Add \$HOME/.local/bin to PATH to access hz-cli from any directory
  To add it to your path, simply execute:
  echo "export PATH=\$HOME/.local/bin:\$PATH" >> \$HOME/.bashrc

EOF

read -rd '' zshAutocompletionDirectives << EOF
* To enable autocompletion capability for Zsh if you have not already
  Append the line below in your .zshrc file if it doesn't exists
  autoload -U compinit; compinit

* Enable autocompletion for Hazelcast Commandline Client (CLC)
  Create a symbolic link of autocompletion script to one of your paths in your fpath such as
  sudo ln -s $HZCLI_HOME/autocompletion/zsh/hz-cli \${fpath[1]}/_hz-cli

* Restart your terminal for the CLC autocompletion to take effect
  or renew your session via:
  /usr/bin/zsh

EOF

read -rd '' bashAutocompletionDirectives << EOF

* Restart your terminal for the CLC autocompletion to take effect
  or renew your session via:
  exec "\$BASH"

EOF

echo "Installation for ZSH:"
if [[ ! -r $HOME/.zshrc || ! "$(cat $HOME/.zshrc)" == *"$(echo "export PATH=$HOME/.local/bin:$PATH")"* ]]; then
    echo "$addToPathDirectivesZSH"
    echo
fi
curl --silent "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/zsh_completion.zsh" --output $HOME/.zsh_completion.sh
mkdir -p $HZCLI_HOME/autocompletion/zsh
mv $HOME/.zsh_completion.sh $HZCLI_HOME/autocompletion/zsh/hz-cli
echo "$zshAutocompletionDirectives"

echo
echo "Installation for Bash:"
if [[ ! $PATH == *"$HOME/.local/bin"* ]]; then
    echo "$addToPathDirectivesBASH"
    echo
fi
xdg_home="$XDG_DATA_HOME"
if [ -z "$xdg_home" ]; then
    # XDG_DATA_HOME was not set
    xdg_home="$HOME/.local/share"
fi
bash_completion_dir="$BASH_COMPLETION_USER_DIR"
if [ -z "$bash_completion_dir" ]; then
    # BASH_COMPLETION_USER_DIR was not set
    bash_completion_dir="$xdg_home/bash-completion"
fi
mkdir -p "${bash_completion_dir}/completions"
mkdir -p "${HZCLI_HOME}/autocompletion/bash"
curl --silent "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/bash_completion.sh" --output "${HZCLI_HOME}/autocompletion/bash/hz-cli"
ln -s $HZCLI_HOME/autocompletion/bash/hz-cli "${bash_completion_dir}/completions/hz-cli"
echo "$bashAutocompletionDirectives"