#!/usr/bin/env bash
if [ "$(uname)" == "Darwin" ]; then
    cat completion_bash.sh > /usr/local/etc/bash_completion.d/hz-cli
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
    cat completion_bash.sh > /etc/bash_completion.d/hz-cli
fi
sleep(1)
echo "Installation completed. For changes to take effect,\nclose the current terminal and reopen."
