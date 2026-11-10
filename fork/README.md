# lazygit (fork)

A fork of [lazygit](https://github.com/jesseduffield/lazygit) aiming at a GitKraken-like
experience out of the box: a colorful commit graph with the branches, remotes and tags labelled
on their commits, without any user configuration.

The `custom` branch (the default) is upstream's latest release plus the changes below; `master`
mirrors upstream. Everything the fork adds for itself lives in `fork/` and
`.github/workflows/fork-release.yml`.

## Changes versus upstream

### Graph and commit labels

- **Commit graph** (`pkg/gui/presentation/lanegraph`): the graph is laid out the way
  GitKraken draws it. Every line of history keeps its column from its tip down to the commit
  where it meets another line, a merge line curves into the line of the branch it merged,
  new lines take the leftmost free column, and lines never shift sideways. Commits are `●`,
  merges `•`; with uncommitted changes the WIP row gets a dotted `◌` node connected to HEAD,
  which keeps the checked-out branch in the leftmost lane. Rows are padded to the graph's
  width so the subjects line up. One deliberate difference from GitKraken: a commit that
  several lines head for goes into the leftmost of them, so the leftmost line (the
  checked-out branch) never hops sideways. Config: `git.log.graphStyle` (`lanes`, the
  default, or `classic` for upstream's author-colored graph). `git.log.order` now defaults to
  `date-order`: it interleaves branches the way GitKraken does and closes merged branches
  quickly, while `topo-order` keeps every merged branch's lane open until its commits are
  listed, which is what made the old graph so wide.
- **Lane colors** (`pkg/gui/presentation/palette`): graph columns are colored cyclically from
  a fixed 10-color palette instead of by commit author. Branch names in the branches panel
  and the status bar use the color of the lane their head commit is on, so a branch always
  matches its line in the graph.
- **Ref labels** (`pkg/gui/presentation/ref_labels.go`): every branch, remote branch and tag
  pointing at a commit is rendered as a colored label on that commit's row, in the commit's
  lane color. The checked-out branch is bold with a `✓`, a branch checked out in another
  worktree gets a `⌂`, remote branches use a darker fill, tags are gold, a detached HEAD is
  red. More than four labels collapse into `+N`. Config: `gui.showRefLabelsInCommitsView`
  (default `true`; `false` restores upstream's `*` marker and plain tags).
- **Typed refs** (`pkg/commands/git_commands/commit_refs.go`): `git log` runs with
  `--decorate=full` and its decorations are parsed into `Commit.Refs` (kind, name, is-HEAD).
  `ExtraInfo` keeps upstream's short form, so search and custom commands are unaffected.
- **Whole graph by default**: `git.log.showWholeGraph` defaults to `true` (`git log --all`),
  with `refs/stash` and `refs/notes/*` excluded since they only add noise.

### Sidebar

- **Branch folders** (`pkg/gui/presentation/ref_groups.go`): local and remote branches whose
  names share a directory-like prefix (`feature/`, `fix/`) are listed together under a folder
  header, shown as the leaf name indented below it, GitKraken-style. Groups keep the position
  of their most recent branch; a prefix used by a single branch is not grouped. Config:
  `gui.groupBranchesByPrefix` (default `true`).
- **Remote branches are colored like their graph lane**, the same way local branches are.
- **Unfolded remotes** (`pkg/gui/presentation/remote_branches.go`): the Remotes tab lists every
  remote's branches directly under a `▾ origin` header, with the same columns as local
  branches (age of the last commit, commit hash, name) and the same prefix folders, instead of
  listing the remotes and showing a remote's branches only after entering one. Checkout, new
  branch, merge, rebase, delete, set upstream and enter (the branch's commits) work as in the
  old remote branches list; `f` fetches the selected branch's remote and `e` opens a remote
  options menu (fetch, edit, remove, add a remote, add a fork remote), which replaces the old
  remotes list. Config: `gui.unfoldRemotes` (default `true`; `false` restores upstream's
  remotes list).
- **Defaults**: commit hash and ahead/behind counts against the base branch are shown in the
  branches panel (`gui.showBranchCommitHash: true`, `gui.showDivergenceFromBaseBranch:
  arrowAndNumber`); Worktrees is a tab of the branches window; the focused side panel expands
  (`gui.expandFocusedSidePanel: true`, weight 3) and the side column is 40% wide.
- **Branch colors**: a branch that matches none of `gui.theme.branchColorPatterns` gets a stable
  palette color instead of the default text color.

### Working tree

- **Staging sections** (`pkg/gui/filetree/sections.go`): the files view lists files with
  unstaged changes first and fully staged files after, under "Unstaged files (n)" and "Staged
  files (m)". A file with both kinds of changes sits in the unstaged section with its `MM`
  status. In the tree view (the default; `~` toggles the flat list) each section is a
  directory tree of its own under a node of its own, so a directory with files in both
  sections appears in both. Both sections are listed whenever there are changes, an empty one
  as "Staged files (0)". Space on a directory stages or unstages the files listed under it,
  and on a section's node all the files of the section; a section collapses like a directory.
  The commands that act on a path (open, ignore, copy the path, the diff tool) are disabled on
  a section's node. Config: `gui.showStagingSectionsInFilesView`.
- **WIP row** (`pkg/gui/presentation/working_tree.go`): when the working tree has changes,
  the commits view starts with `◌ WIP · 1 staged · 2 unstaged · 1 untracked`, like
  GitKraken's work-in-progress node. It re-renders whenever the files refresh. Config:
  `gui.showWorkingTreeInCommitsView`.
- **Stash rows** (`pkg/gui/presentation/stashes_in_graph.go`): the lanes graph shows each
  stash as a `▣` node with a dotted line to the commit it was made on, labelled
  `stash@{n}` and followed by the stash's message, like GitKraken's stash nodes. A stash's row
  goes above the first commit older than the stash, never below the commit it was made on; its
  line takes a column of its own (never the leftmost) and curves into that commit's node
  without taking over the commit's line. The rows can't be selected; the stash panel acts on
  stashes. They are left out while rebasing, filtering, or showing a divergence, and for a
  stash made on a commit that isn't loaded. Config: `gui.showStashesInCommitsView`.

### Main panel

- **Commit header**: a commit's patch view starts with the refs, the subject and body, author
  and committer with human-readable dates, and the parent hashes, followed by git's `--stat`
  summary and the diff. Config: `git.commitDetailsFormat`, a `--pretty` format whose default
  is `DefaultCommitDetailsFormat` in `pkg/config/fork_defaults.go`; empty restores git's
  default header.
- **Log views**: the status panel shows the whole-repo log (`gui.statusPanelView:
  allBranchesLog`), and selecting a branch shows that branch's log.
- **Log graph** (`pkg/gui/controllers/helpers/log_graph_helper.go`): those logs, and the logs
  shown for a selected remote branch or tag, are rendered with the same commit list as the
  commits panel (graph, ref labels, hashes, dates, authors), one commit per line, instead of
  the text output of `git log --graph`. They show the first 300 commits, loaded in the
  background. The status panel's "all branches log" keys cycle from the graph through the text
  formats in `git.allBranchesLogCmds`; a selected branch's text log is `git.branchLogCmd`.
  Config: `gui.mainViewCommitGraph` (default `true`).
- **Pretty diffs** (`pkg/gui/presentation/prettydiff`): diffs in the main panel show the old
  and new line numbers in a gutter, added and removed lines on green and red tinted
  backgrounds with the changed words highlighted, a header per file with its counts of added
  and removed lines and facts such as "new file" or "renamed from", and tinted hunk headers.
  git's output is formatted as it streams in; a commit's header and `--stat` summary pass
  through untouched, and combined diffs of merge conflicts and binary patches are left as git
  prints them. Every row that shows a line of the diff starts with an OSC 1717 record naming
  that line (`records.go`), so the pretty layout stays on screen when the main view is
  focused, and staging, discarding, building a custom patch and copying lines all work on it.
  The formatting is skipped when a custom diff renderer is configured, or git is given
  arguments of its own (`git.diffRenderers`). Config: `gui.prettyDiff` (default `true`) and
  `gui.prettyDiffTheme` (`dark`, the default, or `light`).

### Mouse

- **Text selection**: dragging with the left button in a main panel selects text, drawn in
  reverse video; releasing the button copies it to the clipboard and a toast confirms it,
  `ctrl+o` copies the selection again, and `esc` clears it. This works over anything a main
  panel shows, e.g. to copy a commit's hash from its header. A drag that starts on the gutter
  of a diff's line (a pretty diff's line numbers, or the `+`/`-` column of git's own diff)
  selects diff lines as in upstream instead, e.g. to stage them. Selecting text leaves the
  focus and the selected diff lines alone: a press that starts a text selection acts as a click
  (focusing the panel, selecting the clicked diff line) only when the button is released without
  a drag. Config: `gui.mouseTextSelection` (default `true`).
- **Scrollbars** (`pkg/gocui/scrollbar.go`): dragging a view's scrollbar (the thumb on the
  right edge of its frame) scrolls that view, and pressing the track elsewhere moves the
  thumb's middle there. It scrolls the way the mouse wheel does, without moving the view's
  selection or the focus.
- **Stable half-screen layout**: clicking a main panel (or focusing it with the
  focus-main-view key) in half-screen mode keeps the side panels at their size instead of
  collapsing them; only full-screen mode gives the main panel the whole screen. Config:
  `gui.keepSidePanelsWhenMainFocused` (default `true`).

### Keeping upstream's tests green

`pkg/config/fork_defaults.go` holds `ApplyUpstreamDefaults`, which the integration-test
harness calls so that upstream's integration tests keep running on upstream's defaults; a test
opts into a fork feature in its `SetupConfig` (see `main_view/*_of_a_pretty_diff.go`). Unit
tests for the fork's behavior live next to the code. Upstream tests that assert what the fork
changes on purpose are adjusted or skipped with a reason (`TestGraphColorsFollowTheAuthorColors`).

## Working on it

Everything runs in a container (Docker or OrbStack); no Go toolchain is needed on the host.

```sh
fork/dev test                                   # unit tests
fork/dev test integration                       # all integration tests
fork/dev test integration main_view/stage       # those whose name matches
fork/dev build                                  # fork/dist/lazygit-linux-x64
fork/dev build darwin-arm64                     # fork/dist/lazygit-darwin-arm64
fork/dev shell                                  # a shell with Go, the repo at /work
```

In `fork/dev shell`, run `go generate ./...` after changing config fields or adding an
integration test.

On macOS, delete an existing binary before copying a new build over it (`rm -f` then `cp`):
overwriting a Mach-O file in place invalidates the kernel's cached code signature and the new
binary is killed on launch (exit status 137).

## Releasing

```sh
fork/dev release        # tags release-<upstream tag>, e.g. release-v0.66.2, and pushes it
fork/dev release 2      # a second release on the same upstream tag: release-v0.66.2-2
```

The tag starts `.github/workflows/fork-release.yml`: unit and integration tests, a static Linux
x64 binary and a macOS arm64 binary (cross-compiled, then run on a macOS runner, which the
Release does not wait for since GitHub sometimes has none to give), and a GitHub
Release with `lazygit-<ref>-<target>.tar.gz` files, each holding a bare `lazygit` executable,
and `SHA256SUMS`. Running the workflow by hand builds the binaries as artifacts only.
Upstream's own workflows are kept unchanged, to merge cleanly, and are disabled in the
repository's Actions settings.

## Updating to a newer upstream release

```sh
git fetch upstream --tags
git checkout master && git merge --ff-only upstream/master && git push origin master
git checkout custom && git merge vX.Y.Z      # resolve conflicts, then fork/dev test (both suites)
```

The files most likely to conflict are `pkg/gui/presentation/commits.go`,
`pkg/gui/presentation/branches.go`, `pkg/gui/controllers/main_view_controller.go`,
`pkg/gui/controllers/{branches,status,tags}_controller.go`, `pkg/gui/tasks_adapter.go`,
`pkg/gui/main_panels.go`, `pkg/gui/types/rendering.go`, `pkg/commands/git_commands/commit.go`,
`pkg/commands/git_commands/remote_loader.go`, `pkg/config/user_config.go`, and tests asserting
the git commands lazygit runs (`commit_test.go`, `branch_test.go`); everything else is in new
files. Check `ApplyUpstreamDefaults` against upstream's `GetDefaultConfig` after each merge.
