echo "Installation began."
echo "sudo autoload -U compinit; compinit" >> ~/.zshrc
cat completion_zsh.zsh > "${fpath[1]}/_hz-cli"
sleep 1
echo "Installation completed. For changes to take effect,\nclose the current terminal and reopen."
