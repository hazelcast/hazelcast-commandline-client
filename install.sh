#!/usr/bin/env bash

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
            *"x86_64") bin_id='Linux64_x86_64' ;;
        esac
    ;;
    "darwin")
        case "$machine" in
            *"arm64") bin_id='Darwin_arm64' ;;
            *"x86_64") bin_id='Darwin_x86_64' ;;
        esac
    ;;
esac

tag=$(githubLatestTag hazelcast/hazelcast-commandline-client)
releaseUrl=$(printf "https://github.com/hazelcast/hazelcast-commandline-client/releases/download/v%s/hazelcast-commandline-client_%s_%s.tar.gz" "$tag" "$tag" "$bin_id")

curl -L "$releaseUrl" > "hz-cli_$tag.tar.gz"
tar -xvzf "hz-cli_$tag.tar.gz" "hz-cli_$tag/hz-cli"

mv "hz-cli_$tag/hz-cli" "./hz-cli"
rm -rf "hz-cli_$tag.tar.gz" "hz-cli_$tag/"

case "$(printf "${SHELL##*bin\/}")" in
    "zsh")
        completionUrl=$(printf "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/%s" "zsh_completion.zsh")
        curl -L "$completionUrl" > "./zsh_completion.sh"
        echo "autoload -U compinit; compinit" >> ~/.zshrc;
        cat "./zsh_completion.sh" > "${fpath[1]}/_hz-cli"
    ;;
    "bash")
        completionUrl=$(printf "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/%s" "bash_completion.sh")
        curl -L "$completionUrl" > "./bash_completion.sh"
        if [ "$(uname)" == "Darwin" ]; then
            cat "bash_completion.sh" > /usr/local/etc/bash_completion.d/hz-cli
        elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
            cat "bash_completion.sh" > /etc/bash_completion.d/hz-cli
        fi
    ;;
esac

cat <<-'EOM'

    +       +  o    o     o     o---o o----o o      o---o     o     o----o o--o--o
    + +   + +  |    |    / \       /  |      |     /         / \    |         |
    + + + + +  o----o   o   o     o   o----o |    o         o   o   o----o    |
    + +   + +  |    |  /     \   /    |      |     \       /     \       |    |
    +       +  o    o o       o o---o o----o o----o o---o o       o o----o    o

Hazelcast CommandLine Client is installed.
You can run it with:

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
