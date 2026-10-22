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

A list of Git repositories managed by the content of `gitjoin.txt`. In the above, the context for the `mybranch` branch in `debounce` lives in `mymonorepo/bep/gocode/libs/.mycontext/debounce/mybranch`.

That's it as far as schema goes. We're not enforcing any restrictions on what gets stored inside each context folder, but I'll list some of my imagined uses of this:

* Claude Code can store the current session ID and other relevant information to be used to e.g. restore the session when restarting the work on a feature branch.
* Store documents received from security researchers that describes a security issue to be fixed.
* Store notes/design documents about a given feature.

## Specification

`mycontext` does not depend on `gitjoin`; it works with any Git repository. All commands are run from inside the repository and operate on that repository only.

* The context for a branch lives in `<repo>/../.mycontext/<repo dir name>/<branch>/`.
* Branch names with slashes nest as directories, e.g. `feature/foo` lives in `.mycontext/myrepo/feature/foo/`.
* A context directory is marked by an empty `.mycontext-branch` file. This is how `mycontext` tells a context apart from a sub directory inside one.
* `.mycontext` is meant to be committed (e.g. into the parent `gitjoin` repository). Files that only make sense on one machine go into the context's `.local/` sub directory, which is ignored by `.mycontext/.gitignore`, created by `mycontext` the first time it creates a context.
* `mycontext` never commits anything, and does not touch your Git branches.

## Usage

```bash
mycontext path [--create] [--local]  # Print the context directory for the current branch.
mycontext ls                         # List this repository's contexts; contexts without a local branch are marked (gone).
mycontext rm [branch]                # Remove a context (defaults to the current branch).
mycontext mv <old> <new>             # Rename a context, e.g. after renaming the branch.
mycontext prune [--dry-run]          # Remove contexts for local branches that no longer exist.
```

`mycontext path` is the integration point for other tools, e.g. `cd $(mycontext path --create)`.

### Claude Code

Store the session ID for the current branch with a `SessionStart` [hook](https://docs.claude.com/en/docs/claude-code/hooks) in `.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jq -r .session_id > \"$(mycontext path --create --local)/claude-session\""
          }
        ]
      }
    ]
  }
}
```

And resume it with:

```bash
claude --resume $(cat $(mycontext path --local)/claude-session)
```
