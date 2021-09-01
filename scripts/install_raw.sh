
read -rd '' bashrcAddition << EOF
for bcfile in \$HOME/.bash_completion.d/* ; do
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
        if [ ! -d "$HOME/.local" ]; then
            mkdir $HOME/.local
        fi
        if [ ! -d "$HOME/.local/bin" ]; then
            mkdir $HOME/.local/bin
        fi
        if [[ ! "$(cat $HOME/.zshrc)" == *"$(echo "export PATH=$HOME/.local/bin:$PATH")"* ]]; then
            echo "export PATH=$HOME/.local/bin:$PATH" >> $HOME/.zshrc
        fi
        if [ ! -f "$HOME/.local/bin/hz-cli" ]; then
            mv ./hz-cli $HOME/.local/bin/hz-cli
        fi
        completionUrl=$(printf "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/%s" "zsh_completion.sh")
        curl --silent "$completionUrl" --output $HOME/.zsh_completion.sh
        if [[ ! "$(cat $HOME/.zshrc)" == *"$(echo "autoload -U compinit; compinit")"* ]]; then
            echo "autoload -U compinit; compinit" >> $HOME/.zshrc
        fi
        cat $HOME/.zsh_completion.sh > "${fpath[1]}/_hz-cli"
        rm -rf $HOME/.zsh_completion.sh
    ;;
    "bash")
        if [ ! -d "$HOME/.local" ]; then
            mkdir $HOME/.local
        fi
        if [ ! -d "$HOME/.local/bin" ]; then
            mkdir $HOME/.local/bin
        fi
        if [[ ! "$(cat $HOME/.bashrc)" == *"$(echo "export PATH=$HOME/.local/bin:$PATH")"* ]]; then
            echo "export PATH=$HOME/.local/bin:$PATH" >> $HOME/.bashrc
        fi
        if [ ! -f "$HOME/.local/bin/hz-cli" ]; then
            mv ./hz-cli $HOME/.local/bin/hz-cli
        fi
        if [ ! -d "$HOME/.bash_completion.d" ]; then
            mkdir $HOME/.bash_completion.d
        fi
        completionUrl=$(printf "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/%s" "bash_completion.sh")
        curl --silent "$completionUrl" --output $HOME/.bash_completion.d/hz-cli
        if [[ ! "$(cat $HOME/.bashrc)" == *"$(echo "$bashrcAddition")"* ]]; then
            echo "$bashrcAddition" >> $HOME/.bashrc
        fi
        source $HOME/.bashrc
    ;;
esac

clear

echo "$greetings"
while :
do
    echo "$install_options"
    read selection
    case "$selection" in
        "1")
            clear;
            echo "$local_manual";
            break;
        ;;
        "2")
            clear;
            echo "$cloud_manual";
            break;
        ;;
        "3")
            clear;
            exit 0;
        ;;
        *)
            clear;
            echo "Unknown option. Try again.\n";
    esac
done

case "$(printf "${SHELL##*bin\/}")" in
    "zsh")
        /bin/zsh;
    ;;
    "bash")
        /bin/bash;
    ;;
esac
