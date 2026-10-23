I manage my Git repos with [gitjoin](https://github.com/bep/gitjoin) together with other small tools such as [firstupdotenv](https://github.com/bep/firstupdotenv) and [grrep](https://github.com/bep/grrep), each doing one thing very well. What I missed was a simple way to keep contextual information per repository and branch. The tools I've seen that do this look like sledgehammers for my nails, so I wrote `mycontext`.

In my setup, `mymonorepo/bep/gocode/libs` is a `gitjoin` virtual monorepo with Git repos such as `debounce`, `gitmap` and `lazycache`. The context for the `mybranch` branch in `debounce` lives in `libs/_mycontext/debounce/mybranch/`.

What goes into a context is up to you. Some examples:

* The current Claude Code session ID, to resume the session when picking up work on a branch (see below).
* Reports from security researchers describing an issue to be fixed.
* Notes and design documents for a feature.

## Specification

`mycontext` does not depend on `gitjoin`; it works with any Git repository. All commands are run from inside the repository and operate on that repository only.

* The context for a branch lives in `<repo>/../_mycontext/<repo dir name>/<branch>/`.
* Branch names with slashes nest as directories, e.g. `feature/foo` lives in `_mycontext/myrepo/feature/foo/`.
* A context directory is marked by an empty `.mycontext-branch` file. This is how `mycontext` tells a context apart from a sub directory inside one.
* `_mycontext` is meant to be committed (e.g. into the parent `gitjoin` repository). Files that only make sense on one machine go into the context's `.local/` sub directory, which is ignored by `_mycontext/.gitignore`, created by `mycontext` the first time it creates a context.
* `mycontext` never commits anything, and does not touch your Git branches.

## Installation

```bash
go install github.com/bep/mycontext@latest
```

## Usage

```bash
mycontext path [--local]             # Print the context directory for the current branch, creating it if needed.
mycontext ls                         # List this repository's contexts; contexts without a local branch are marked (gone).
mycontext rm [branch]                # Remove a context (defaults to the current branch).
mycontext mv <old> <new>             # Rename a context, e.g. after renaming the branch.
mycontext prune [--dry-run]          # Remove contexts for local branches that no longer exist.
```

`mycontext path` is the integration point for other tools, e.g. `cd $(mycontext path)`.

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
            "command": "jq -r .session_id > \"$(mycontext path --local)/claude-session\""
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
