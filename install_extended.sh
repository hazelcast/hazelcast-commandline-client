#!/usr/bin/env bash

read -rd '' greetings << EOF

+       +  o    o     o     o---o o----o o      o---o     o     o----o o--o--o
+ +   + +  |    |    / \       /  |      |     /         / \    |         |
+ + + + +  o----o   o   o     o   o----o |    o         o   o   o----o    |
+ +   + +  |    |  /     \   /    |      |     \       /     \       |    |
+       +  o    o o       o o---o o----o o----o o---o o       o o----o    o

Hazelcast CommandLine Client is installed.
You can run it with:

./hz-cli

If you already have a cluster running, you can start with putting and retrieving an entry in Map:

   ./hz-cli map --name myMap put --key 1 --value "Hello"
   ./hz-cli map --name myMap get --key 1
   > Hello

Tip: See --address, --cloud-token and --cluster-name for connecting to a non-local default cluster.

EOF

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

clear;

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
finalUrl=$(printf "https://github.com/hazelcast/hazelcast-commandline-client/releases/download/v%s/hazelcast-commandline-client_%s_%s.tar.gz" "$tag" "$tag" "$bin_id")

curl -L "$finalUrl" > "hz-cli_$tag.tar.gz"
tar -xvzf "hz-cli_$tag.tar.gz" "hz-cli_$tag/hz-cli"

mv "hz-cli_$tag/hz-cli" "./hz-cli"
rm -rf "hz-cli_$tag.tar.gz" "hz-cli_$tag/"

case "$(printf "${SHELL##*bin\/}")" in
    "zsh")
        completionUrl=$(printf "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/%s" "zsh_completion.sh")
        curl "$completionUrl" --output "./zsh_completion.sh"
        echo "autoload -U compinit; compinit" >> ~/.zshrc;
        cat "./zsh_completion.sh" > "${fpath[1]}/_hz-cli"
    ;;
    "bash")
        completionUrl=$(printf "https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/extras/%s" "bash_completion.sh")
        curl "$completionUrl" --output "./bash_completion.sh"
        if [ "$(uname)" == "Darwin" ]; then
            cat "./bash_completion.sh" > /usr/local/etc/bash_completion.d/hz-cli
        elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
            cat "./bash_completion.sh" > /etc/bash_completion.d/hz-cli
        fi
    ;;
esac

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
