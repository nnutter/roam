---
name: roam
description: Version-control dotfiles with roam, a Git wrapper for a home-directory work tree.
---

# roam

Use `roam` instead of Git to version-control the user's dotfiles when the files are not in a Git repository.
`roam` runs Git against a bare repository and a home-directory work tree, so one repository can track files from many directories under `$HOME`.

## When to use

- A dotfile needs version control, and no Git repository already contains it.
- The user asks to set up dotfile tracking on a new host or account.
- The user asks to update, push, or inspect the dotfiles repository.

## Commands

- `roam setup -r <repo> -b <branch>` initializes the bare repository and the home work tree.
- `roam update` fetches and rebases with autostash.
- `roam sync` updates, then pushes.
- Every other argument passes to Git, as in `roam status --untracked-files -- ~/bin`.

Set `status.showUntrackedFiles no` in the repository, so `roam status` reports only tracked files.
