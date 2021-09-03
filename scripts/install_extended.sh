#!/usr/bin/env bash

read -rd '' local_manual << EOF

* Local Cluster Manual *

1. Install Hazelcast through 'hz' command line tool

   brew tap hazelcast/hz
   brew install hazelcast
   
   Tip: No using Homebrew? See https://docs.hazelcast.com/hazelcast/latest/getting-started/install-hazelcast.html for other installation options.
   
2. Start Hazelcast member locally

   hz start
   
3. You have all you need. Put and retrieve your first entry in Hazelcast:

    ./hz-cli map --name myMap put --key 1 --value "Hello"
    ./hz-cli map --name myMap get --key 1
    > Hello     

4. Enjoy! Visit https://docs.hazelcast.com for more information on Hazelcast!
EOF

read -rd '' cloud_manual << EOF

* Cloud Cluster Manual *

1. Go to https://cloud.hazelcast.com and login/create an account
2. In the Cloud console, go to Account -> Developer and write down the API Key and API Secret.
3. Install 'hzcloud' command line tool

   brew tap hazelcast/hz
   brew install hzcloud
   
   Tip: No using Homebrew? See https://github.com/hazelcast/hazelcast-cloud-cli#installing-hzcloud for other installation options.
   
4. Login to your Hazelcast Cloud account with the credentials retrieved above:

   hzcloud login

5. Create a cluster:

   hzcloud starter-cluster create \
   --cloud-provider=aws \
   --cluster-type=FREE \
   --name=mycluster \
   --region=us-east-1 \
   --total-memory=0.2 \
   --hazelcast-version=4.0
  
6. Wait till the cluster state is "RUNNING". Make note of the cluster ID.

   hzcloud starter-cluster list

7. Retrive the discovery token (Discovery Tokens -> Token) and cluster name (\"pr-<ID of the cluster>\", currently a workaround for a bug in Hazelcast Cloud):

   hzcloud starter-cluster get --cluster-id <ID of the cluster from previous step>
   
8. You have all you need. Put and retrieve your first entry in Hazelcast Cloud:

   ./hz-cli --cloud-token <token from previous step> --cluster-name <cluster name from previous step> map --name myMap put --key 1 --value "Hello"
   ./hz-cli --cloud-token <token from previous step> --cluster-name <cluster name from previous step> map --name myMap get --key 1
   > Hello
    
9. Enjoy! Visit https://docs.hazelcast.com for more information on Hazelcast Cloud and Hazelcast configuration!
EOF

read -rd '' install_options << EOF
If you don't have a running cluster, would you like to see a tutorial how to do it? [select option]
1. Yes, I want to setup a locally running cluster.
2. Yes, I want to setup a cluster in a managed service Hazelcast Cloud.
3. No, I have all I need, I want to exit the installation script.

Select option 1, 2 or 3 and press Enter:
EOF

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

selectHzManual() {
    while :
    do
        echo "$install_options"
        read selection
        case "$selection" in
            "1")
                echo "$local_manual";
                break;
            ;;
            "2")
                echo "$cloud_manual";
                break;
            ;;
            "3")
                exit 0;
            ;;
            *)
                echo "Unknown option. Try again.\n";
        esac
    done
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
                echo "* Add path automatically?(yes/no/exit)"
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
                        echo "You can do this manually by"
                        echo "Opening your \$HOME/.zshrc file and appending the line below:"
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
                echo "* Would you like to enable autocompletion capability automatically?(yes/no/exit)"
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
                        echo "Autocompletion capability can be enabled by appending into the \$HOME/.zshrc the line below:"
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
                echo "* Would you like to enable CLC autocompletion?(yes/no/exit)"
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
            echo "* Would you like to renew your session now or later?(now/later/exit)"
            read answer
            case "$answer" in
                "now")
                    echo
                    echo "Session is renewed."
                    echo
                    selectHzManual
                    /usr/bin/zsh
                    break
                ;;
                "later")
                    echo
                    echo "Don't forget to renew your session later."
                    echo
                    selectHzManual
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
            echo "* Would you like the installation script to handle this?"
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
                        echo "You can do this manually by"
                        echo "Opening your \$HOME/.bashrc file and appending the line below:"
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
                echo "* Would you like to enable CLC autocompletion?(yes/no/exit)"
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
            echo "* Would you like to renew your session now or later?(now/later/exit)"
            read answer
            case "$answer" in
                "now")
                    echo
                    echo "Session is renewed."
                    echo
                    selectHzManual
                    /usr/bin/bash
                    break
                ;;
                "later")
                    echo
                    echo "Don't forget to renew your session later."
                    echo
                    selectHzManual
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
