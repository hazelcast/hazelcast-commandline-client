#!/usr/bin/env bash
git clone https://github.com/hazelcast/hazelcast-commandline-client.git
cd hazelcast-commandline-client
go build -o hz-cli github.com/hazelcast/hazelcast-commandline-client
if [ "$(uname)" == "Darwin" ]; then
    completion_bash.sh > /usr/local/etc/bash_completion.d/hz-cli
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
    completion_bash.sh > /etc/bash_completion.d/hz-cli
fi
sleep(1)
echo "Installation completed. For changes to take effect,\nclose the current terminal and reopen."
