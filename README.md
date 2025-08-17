# roam

Way back when I realized I could reasonably manage my dotfiles just using `git` by setting `status.showUntrackedFiles no`.
One caveat is that it's a little hard to tell what is or isn't managed.
However, it's not too hard to check specific locations, e.g.,

```
roam status --untracked-files -- ~/bin
```

You can also still use `.gitignore` to ignore files you definitely do not want to ever manage.

This has served me well for over 15 years.
My one slight annoyance is the manual set up required for a new host/account which this repo and this version of `roam` will address.
