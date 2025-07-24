#!/usr/bin/env bash

# Original notes on this are at https://gist.github.com/nnutter/42a83f2a0479c9a53b813b3c49eff09d.
#
# TODO:
#   - Add help docs.
#   - Add subcommand to install roam.sh permanently.
#   - Add subcommand to complete roam as git for bash and zsh:
#       - `complete -F _git roam` for bash. In the past it looks like I also used
#           the options, `-o bashdefault -o default -o nospace`.
#       - `compdef roam=git` for zsh

_roam() {
	git --git-dir "$HOME/.dotfiles.git" --work-tree "$HOME" "$@"
}

help() {
	echo "1. ./roam.sh setup"
	echo "2. ./roam.sh checkout master"
	echo "3. PROFIT"
}

setup-init() {
	_roam init
	_roam config status.showUntrackedFiles no
}

setup-remote() {
	_roam remote add -f origin git@github.com:"$1".git
}

setup() {
	setup-init
	setup-remote "$@"
}

if test "$#" -eq 0
then
	help
	exit 1
fi

case "$1" in
help)
	help "$@";;
setup)
	setup "$@";;
setup-init)
	setup-init "$@";;
setup-remote)
	setup-remote "$@";;
*)
	_roam "$@"
esac
