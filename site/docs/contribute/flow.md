# GitHub Workflow

This page describes the GitHub workflow that contributors should follow.

## Issues and pull requests

Contributing to Mesh for Data is done following the GitHub workflow of Pull Requests.

You should usually open a pull request in the following situations:

- Start work on a contribution that was that you’ve already discussed in an issue.
- Submit trivial fixes (for example, a typo, a broken link or an obvious error).

A pull request doesn’t have to represent finished work. It’s usually better to open a draft pull request early on, so others can watch or give feedback on your progress.

Here’s how to submit a pull request:

- **[Fork](https://github.com/ibm/the-mesh-for-data/fork)** the main repository
- **Clone the forked repository locally**. Connect your local to the original “upstream” repository by adding it as a remote.
    ```shell
    git clone git@github.com:$(git config user.name)/the-mesh-for-data.git
    git remote add upstream https://github.com/ibm/the-mesh-for-data.git
    git remote set-url --push upstream no_push
    ```
- **[Pull in changes](https://help.github.com/articles/syncing-a-fork/)** from “upstream” often so that you stay up to date so that when you submit your pull request, merge conflicts will be less likely.
    ```shell
    git fetch upstream master
    git checkout master
    git merge upstream/master
    git push origin master
    ```
- **[Create a branch](https://guides.github.com/introduction/flow/)** for your edits from master. Note that your should never add edits to the master branch itself.
    ```shell
    git checkout -b <branch name>
    ```
- **Make commits of logical units**, ensuring that commit messages are in the [proper format](#format-of-the-commit-message).
- **Push your changes** to the created branch in your fork of the repository.
- **Open a pull request** to the original repository.
- **Reference any relevant issues** or supporting documentation in your PR (for example, “Closes #37.”)

As always, you must [follow code style](#normalize-the-code), ensure that [all tests pass](build-test.md), and add any new tests as appropriate.

**Thanks for your contribution!**

## Normalize the code

To ensure the code is formatted uniformly we use various linters which are
invoked using

```bash
make verify
```

## Format of the Commit Message

The project follows a rough convention for commit messages that is designed to answer two questions: what changed and why.
The subject line should feature the what and the body of the commit should describe the why.

Every commit must also include a DCO Sign Off at the end of the commit message. By doing this you state that you certify the [Developer Certificate of Origin](https://developercertificate.org/). This can be automated by adding the `-s` flag to `git commit`. You can also mass sign-off a whole PR with `git rebase --signoff master`.

Example commit message:
```
scripts: add the test-cluster command

this uses tmux to setup a test cluster that you can easily kill and
start for debugging.

Fixes #38

Signed-off-by: Legal Name <your.email@example.com>
```

The format can be described more formally as follows:

```
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
<BLANK LINE>
<signoff>
```

The first line is the subject and should be no longer than 70 characters, the second line is always blank, and other lines should be wrapped at 80 characters.
This allows the message to be easier to read on GitHub as well as in various git tools.
