#!/usr/bin/env bash

read -rd '' bashrcAddition << EOF
for bcfile in ~/.bash_completion.d/* ; do
    [ -f "\$bcfile" ] && . "\$bcfile"
done
EOF

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

curl -L --silent "$releaseUrl" --output "hz-cli"
chmod +x "./hz-cli"

case "$(printf "${SHELL##*bin\/}")" in
    "zsh")
        completionUrl=$(printf "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/%s" "zsh_completion.sh")
        curl --silent "$completionUrl" --output ~/.zsh_completion.sh
        if [[ ! "$(cat ~/.zshrc)" == *"$(echo "autoload -U compinit; compinit")"* ]]; then
            echo "autoload -U compinit; compinit" >> ~/.zshrc
        fi
        cat ~/.zsh_completion.sh > "${fpath[1]}/_hz-cli"
    ;;
    "bash")
        if [ ! -d "~/.bash_completion.d" ]; then
            mkdir ~/.bash_completion.d
        fi
        completionUrl=$(printf "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/%s" "bash_completion.sh")
        curl --silent "$completionUrl" --output ~/.bash_completion.d/hz-cli
        if [[ ! "$(cat ~/.bashrc)" == *"$(echo "$bashrcAddition")"* ]]; then
            echo "$bashrcAddition" >> ~/.bashrc
        fi
    ;;
esac

clear

cat <<-'EOM'

    +       +  o    o     o     o---o o----o o      o---o     o     o----o o--o--o
    + +   + +  |    |    / \       /  |      |     /         / \    |         |
    + + + + +  o----o   o   o     o   o----o |    o         o   o   o----o    |
    + +   + +  |    |  /     \   /    |      |     \       /     \       |    |
    +       +  o    o o       o o---o o----o o----o o---o o       o o----o    o

Hazelcast CommandLine Client is installed.
For changes to take effect, restart your shell session/terminal.
After that, you can run it with:

./hz-cli

A Simple Guideline:

Q: Do you have a Hazelcast cluster running?
A: If yes, then you can start with ./hz-cli help to learn the tool.
   If no, download & install the Hazelcast via package manager(brew, apt)
   or retrieve the binary from the website.

Q: Do you want to manage your cluster?:
A: If yes, you can use ./hz-cli cluster --help to see available commands.

Q: Do you want to store or retrieve data from a map in your cluster?
A: If yes, you can use ./hz-cli map --help to see available commands.

For any other questions, you can use 

./hz-cli help 

to learn the tool.

EOM
