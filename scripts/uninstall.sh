#!/usr/bin/env bash

PROGRAM_NAME="hzc"
HZCLI_HOME="$HOME/.local/share/hz-cli"

rm -f $HOME/.local/bin/$PROGRAM_NAME
if [[ -r $HOME/.zshrc ]]; then
    echo "You can remove ~/.local/bin from your path by removing a line in the form \"PATH=\$HOME/.local/bin:\$PATH\" from ~/.zshrc"
    echo
fi
if [[ -r $HOME/.bashrc ]]; then
    echo "You can remove ~/.local/bin from your path by removing a line in the form \"PATH=\$HOME/.local/bin:\$PATH\" from ~/.bashrc"
    echo
fi

rm -rf ${HZCLI_HOME} #remove hzcli related files (config.yaml,.hzc.yaml,autocompletion scripts)

echo "To remove autocompletion on Zsh, execute:"
echo "sudo rm \${fpath[1]}/_$PROGRAM_NAME"

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

#remove bash completion link
rm -f "${bash_completion_dir}/completions/$PROGRAM_NAME"
