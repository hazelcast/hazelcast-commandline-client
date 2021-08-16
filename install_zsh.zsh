#!/usr/local/bin/zsh
git clone https://github.com/hazelcast/hazelcast-commandline-client.git
cd hazelcast-commandline-client
go build -o hz-cli github.com/hazelcast/hazelcast-commandline-client
echo "autoload -U compinit; compinit" >> ~/.zshrc
completion_zsh.zsh > "${fpath[1]}/_hz-cli"
sleep(1)
echo "Installation completed. For changes to take effect,\nclose the current terminal and reopen."
