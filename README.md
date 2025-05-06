# Devplan CLI

A command-line interface for Devplan that helps automate development workflows.

## Set up


### Private Github repo

We depend on github.com/devplaninc/webapp, which is a private repo. Foo that to work:

1. Add following into your `~/.gitconfig` file:
```
[url "git@github.com:devplaninc/webapp.git"]
 insteadOf = https://github.com/devplaninc/webapp
```
2. Set up repo
```
export GOPRIVATE=github.com/devplaninc/webapp
go mod tidy
```
