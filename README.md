I manage my Git repos with [GitJoin](https://github.com/bep/gitjoin) in combo with other small tools such as [firstupenv](https://github.com/bep/firstupdotenv) and [grrep](https://github.com/bep/grrep), tools sharing one great common feature: They are small tools each doing one thing very well.

The above has served me very well and raised the level from the slightly chaotic work flow I had before. There's one thing I miss, though, and that is a better way to share per repository/branch contextual information. I have seen more advanced tools in my Twitter feed doing some of this, but they all look a little like sledge hammers for my nails. I could possibly also build this into `gitjoin`, but that would somehow break the "doing one thing very well".

So, this tool, `mycontext`, is a very simple specification and a CLI tool that can be used to manage the different contexts.

In my use, I have a big virtual monorepo managed by `gitjoin`. The content of `mymonorepo/bep/gocode/libs` are:

```bash
awscreate           gitmap              gojap               helpers             overlayfs           textandbinarywriter
debounce            gobench             golibsass           imagemeta           predicate           tmc
execrpc             godartsass          golibtemplate       lazycache           semverpair          triebytesmapper
gitjoin.txt         gogitignore         golocales           logg                simplecobra
```

A list of Git repositories managed by the content of `gitjoin.txt`. In the above, we could store per repository/branch context on the form `mymonorepo/bep/gocode/libs/.mycontext/debounce/mybranch`.

That's it as far as schema goes. We're not enforcing any restrictions on what gets stored inside each context folder, but I'll list some of my imagined uses of this:

* Claude Code can store the current session ID and other relevant information to be used to e.g. restore the session when restarting the work on a feature branch.
* Store documents received from security researchers that describes a security issue to be fixed.
* Store notes/design documents about a given feature.

The CLI would be simple, we would need commands that:

* Creates a new branch in a named repo, e.g. `mycontext new debounce mybranch`; I think it makes sense to also checkout (and create if it doesn't exist) the given branch in that repo.
* Removes a branch (but leave the Git branch alone)
* Prunes contexts for non-existing branches.

