.PHONY: install

install:
case $(echo $SHELL) in 
	*/zsh) zsh install_zsh.zsh;;
	*/bash) sh install_bash.sh;;
esac