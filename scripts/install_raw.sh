#!/usr/bin/env bash

read -rd '' greetings << EOF
==============================================================================

+       +  o    o     o     o---o o----o o      o---o     o     o----o o--o--o
+ +   + +  |    |    / \       /  |      |     /         / \    |         |
+ + + + +  o----o   o   o     o   o----o |    o         o   o   o----o    |
+ +   + +  |    |  /     \   /    |      |     \       /     \       |    |
+       +  o    o o       o o---o o----o o----o o---o o       o o----o    o

Hazelcast CommandLine Client is installed.
You can run it with:

hz-cli

If you already have a cluster running, you can start with putting and retrieving an entry in Map:

   hz-cli map --name myMap put --key 1 --value "Hello"
   hz-cli map --name myMap get --key 1
   > Hello

Tip: See --address, --cloud-token and --cluster-name for connecting to a non-local default cluster.

EOF

read -rd '' zshInstallationManual << EOF

Hazelcast Commandline Client (CLC) Installation Manual:

* Add \$HOME/.local/bin into your \$PATH
- If there's no such folder, you can create it with:
> mkdir -p \$HOME/.local/bin

* Move hz-cli into \$HOME/.local/bin
> mv \$HOME/hz-cli \$HOME/.local/bin/hz-cli

* Download the CLC autocompletion script for Zsh
> curl --silent "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/zsh_completion.sh" --output \$HOME/_hz-cli

* Enable autocompletion capability for Zsh
- Append the line below in your .zshrc file if not exists
> 
autoload -U compinit; compinit

* Enable autocompletion for Hazelcast Commandline Client (CLC)
- Move the CLC autocompletion script to the required place
> mv \$HOME/_hz-cli > /usr/local/share/zsh/site-functions/_hz-cli

* Restart your terminal for the CLC autocompletion to take effect
or renew your session via:
> /usr/bin/zsh

EOF

read -rd '' bashInstallationManual << EOF

Hazelcast Commandline Client (CLC) Installation Manual:

* Add \$HOME/.local/bin into your \$PATH
- If there's no such folder, you can create it with:
> mkdir -p \$HOME/.local/bin

* Move hz-cli into \$HOME/.local/bin
> mv \$HOME/hz-cli \$HOME/.local/bin/hz-cli

* Create user defined bash_completion.d folder
> mkdir -p \$HOME/.bash_completion.d

* Download the CLC autocompletion script for Bash
> curl --silent "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/bash_completion.sh" --output \$HOME/.bash_completion.d/hz-cli

* Enable autocompletion for Hazelcast Commandline Client (CLC)
- Append the lines below to the \$HOME/.bashrc file if not exists
> 
for bcfile in \$HOME/.bash_completion.d/* ; do
    [ -f "\$bcfile" ] && . "\$bcfile"
done

* Restart your terminal for the CLC autocompletion to take effect
or renew your session via:
> /usr/bin/bash

EOF

read -rd '' bashrcAutoCompletionAddition << EOF
for bcfile in \$HOME/.bash_completion.d/* ; do
    [ -f "\$bcfile" ] && . "\$bcfile"
done
EOF

read -rd '' bashrcPathAddition << EOF
echo "export PATH=\$HOME/.local/bin:\$PATH" >> \$HOME/.bashrc
EOF

read -rd '' zshrcAutoCompletionEnabling << EOF
echo "autoload -U compinit; compinit" >> \$HOME/.zshrc
EOF

read -rd '' zshrcAutoCompletionAddition << EOF
mv \$HOME/_hz-cli > /usr/local/share/zsh/site-functions/_hz-cli
EOF

read -rd '' zshrcPathAddition << EOF
echo "export PATH=\$HOME/.local/bin:\$PATH" >> \$HOME/.zshrc
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

echo
if [[ ! -f "$HOME/.local/bin/hz-cli" ]]; then
    echo "Hazelcast Commandline Client is being downloaded."

    curl -L --silent "$releaseUrl" --output "$HOME/hz-cli"
    chmod +x $HOME/hz-cli
    mkdir -p $HOME/.local/bin
    mv $HOME/hz-cli $HOME/.local/bin/hz-cli
    echo "Hazelcast Commandline Client (CLC) is downloaded and placed in \$HOME/.local/bin/hz-cli"
else
    echo "Hazelcast Commandline Client is already downloaded."
    echo "It exists as \$HOME/.local/bin/hz-cli"
    echo
    while :
    do
        echo "* Would you like to redownload Hazelcast Commandline Client?(yes/no/exit)"
        read answer
        case "$answer" in
            "yes")
                echo "Hazelcast Commandline Client is being downloaded."
                curl -L --silent "$releaseUrl" --output $HOME/hz-cli
                chmod +x $HOME/hz-cli
                mkdir -p $HOME/.local/bin
                mv $HOME/hz-cli $HOME/.local/bin/hz-cli
                echo "Hazelcast Commandline Client (CLC) is downloaded and placed as \$HOME/.local/bin/hz-cli"
                break
            ;;
            "no")
                echo
                echo "You can find it as \$HOME/.local/bin/hz-cli"
                break

            ;;
            "exit")
                exit 0
            ;;
            *)
                printf "ERROR: invalid answer: %s" "$answer"
        esac
    done
fi

echo
echo "Hazelcast Commandline Client is being installed."
echo "You may perform your installation either by the installation script or by yourself."
echo
while :
do
    echo "* Would you like to perform installation via script?(yes/no/exit)"
    read answer
    case "$answer" in
        "yes")
            echo
            echo "Installation script is selected."
            echo
            break
        ;;
        "no")
            echo
            case "$(printf "${SHELL##*bin\/}")" in
                "zsh")
                    echo
                    echo "$zshInstallationManual"
                    echo
                    exit 0
                ;;
                "bash")
                    echo
                    echo "$bashInstallationManual"
                    echo
                    exit 0
                ;;
            esac
        ;;
        "exit")
            exit 0
        ;;
        *)
            printf "ERROR: invalid answer: %s" "$answer"
    esac
done

case "$(printf "${SHELL##*bin\/}")" in
    "zsh")
        if [[ ! "$(cat $HOME/.zshrc)" == *"$(echo "export PATH=\$HOME/.local/bin:\$PATH")"* ]]; then
            echo "\$HOME/.local/bin is not added into \$PATH"
            echo "Add $HOME/.local/bin to \$PATH to access Hazelcast Commandline Client from any directory"
            echo "Would you like the installation script to handle this?"
            echo
            while :
            do
                echo "Add path automatically?(yes/no/exit)"
                read answer
                case "$answer" in
                    "yes")
                        echo
                        echo "export PATH=\$HOME/.local/bin:\$PATH" >> $HOME/.zshrc
                        echo "\$HOME/.local/bin is added into \$PATH"
                        break
                    ;;
                    "no")
                        echo
                        echo "you can do this manually by"
                        echo "opening your \$HOME/.zshrc file and appending the line below:"
                        echo "$zshrcPathAddition"
                        break
                    ;;
                    "exit")
                        exit 0
                    ;;
                    *)
                        echo
                        printf "ERROR: invalid answer: %s" "$answer"
                esac
            done
            echo
        fi
        completionUrl=$(printf "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/%s" "zsh_completion.zsh")
        curl --silent "$completionUrl" --output $HOME/_hz-cli
        chmod +x $HOME/_hz-cli
        if [[ ! "$(cat $HOME/.zshrc)" == *"$(echo "autoload -U compinit; compinit")"* ]]; then
            echo "Autocompletion capability is not enabled in this Zsh."
            echo "Would you like the installation script to enable autocompletion capability?"
            echo
            while :
            do
                echo "Enable autocompletion capability automatically?(yes/no/exit)"
                read answer
                case "$answer" in
                    "yes")
                        echo
                        echo "autoload -U compinit; compinit" >> $HOME/.zshrc
                        echo "Autocompletion capability is enabled."
                        break
                    ;;
                    "no")
                        echo
                        echo "autocompletion capability can be enabled by appending into the \$HOME/.zshrc the line below:"
                        echo "$zshrcAutoCompletionEnabling"
                        break
                    ;;
                    "exit")
                        exit 0
                    ;;
                    *)
                        echo
                        printf "ERROR: invalid answer: %s" "$answer"
                esac
            done
            echo
        fi
        if [[ ! -f "/usr/local/share/zsh/site-functions/_hz-cli" ]]; then
            echo "Hazelcast Commandline Client (CLC) autocompletion is not set for Zsh."
            echo "Would you like the installation script to enable CLC autocompletion?"
            echo "If yes, it will require sudo permission to perform this task:"
            echo "$zshrcAutoCompletionAddition"
            echo
            while :
            do
                echo "Enable CLC autocompletion?(yes/no/exit)"
                read answer
                case "$answer" in
                    "yes")
                        echo
                        sudo mv $HOME/_hz-cli /usr/local/share/zsh/site-functions/_hz-cli
                        echo "CLC autocompletion enabled."
                        break
                    ;;
                    "no")
                        echo
                        echo "CLC autocompletion can be enabled by executing the line below for once:"
                        echo "$zshrcAutoCompletionAddition"
                        break
                    ;;
                    "exit")
                        exit 0
                    ;;
                    *)
                        echo
                        printf "ERROR: invalid answer: %s" "$answer"
                esac
            done
            echo
        fi
        echo "$greetings"
        echo "Hazelcast Commandline Client (CLC) requires a new session for the autocompletion to take effect"
        echo "Either restarting the termainal or executing '/usr/bin/zsh' is fine."
        echo
        while :
        do
            echo "Would you like to renew your session now or later?(now/later/exit)"
            read answer
            case "$answer" in
                "now")
                    echo
                    echo "Session is renewed."
                    /usr/bin/zsh
                    break
                ;;
                "later")
                    echo
                    echo "Don't forget to renew your session later."
                    break
                ;;
                "exit")
                    exit 0
                ;;
                *)
                    echo
                    printf "ERROR: invalid answer: %s" "$answer"
            esac
        done
    ;;
    "bash")
        if [[ ! "$(cat $HOME/.bashrc)" == *"$(echo "export PATH=\$HOME/.local/bin:\$PATH")"* ]]; then
            echo "\$HOME/.local/bin is not added into \$PATH"
            echo "Add $HOME/.local/bin to \$PATH to access Hazelcast Commandline Client from any directory"
            echo "Would you like the installation script to handle this?"
            echo
            while :
            do
                echo "Add path automatically?(yes/no/exit)"
                read answer
                case "$answer" in
                    "yes")
                        echo
                        echo "export PATH=\$HOME/.local/bin:\$PATH" >> $HOME/.bashrc
                        echo "\$HOME/.local/bin is added into \$PATH"
                        break
                    ;;
                    "no")
                        echo
                        echo "you can do this manually by"
                        echo "opening your \$HOME/.bashrc file and appending the line below:"
                        echo "$bashrcPathAddition"
                        break
                    ;;
                    "exit")
                        exit 0
                    ;;
                    *)
                        echo
                        printf "ERROR: invalid answer: %s" "$answer"
                esac
            done
            echo
        fi
        if [ ! -d "$HOME/.bash_completion.d" ]; then
            mkdir $HOME/.bash_completion.d
        fi
        completionUrl=$(printf "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/%s" "bash_completion.sh")
        curl --silent "$completionUrl" --output $HOME/.bash_completion.d/hz-cli
        if [[ ! "$(cat $HOME/.bashrc)" == *"$(echo "$bashrcAutoCompletionAddition")"* ]]; then
            echo "Hazelcast Commandline Client (CLC) autocompletion is not set for Bash."
            echo "Would you like the installation script to enable CLC autocompletion?"
            echo
            while :
            do
                echo "Enable CLC autocompletion?(yes/no/exit)"
                read answer
                case "$answer" in
                    "yes")
                        echo
                        echo "$bashrcAutoCompletionAddition" >> $HOME/.bashrc
                        echo "CLC autocompletion enabled."
                        break
                    ;;
                    "no")
                        echo
                        echo "CLC autocompletion can be enabled by adding the section below in the \$HOME/.bashrc file:"
                        echo "$bashrcAutoCompletionAddition"
                        break
                    ;;
                    "exit")
                        exit 0
                    ;;
                    *)
                        echo
                        printf "ERROR: invalid answer: %s" "$answer"
                esac
            done
            echo
        fi
        echo
        echo "$greetings"
        echo "Hazelcast Commandline Client (CLC) requires a new session for the autocompletion to take effect"
        echo "Either restarting the termainal or executing '/usr/bin/bash' is fine."
        echo
        while :
        do
            echo "Would you like to renew your session now or later?(now/later/exit)"
            read answer
            case "$answer" in
                "now")
                    echo
                    echo "Session is renewed."
                    /usr/bin/bash
                    break
                ;;
                "later")
                    echo
                    echo "Don't forget to renew your session later."
                    break
                ;;
                "exit")
                    exit 0
                ;;
                *)
                    echo
                    printf "ERROR: invalid answer: %s" "$answer"
            esac
        done
    ;;
esac
