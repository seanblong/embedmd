# Contributing

Contributions are welcome via GitHub Pull Requests. This document outlines the
process to help get your contribution accepted.

Any type of contribution is welcome; from new features, bug fixes, tests, and documentation.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How to Contribute](#how-to-contribute)
- [Testing](#testing)
   - [Unit Tests](#unit-tests)
   - [Pre-commit](#pre-commit)
      - [Prerequisites](#prerequisites)
      - [Running pre-commit Manually](#running-pre-commit-manually)
- [Opening PRs](#opening-prs)
   - [PR title](#pr-title)

## Code of Conduct

See the [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

1. Fork the repository.
1. Develop your changes on your fork.
1. Run [tests and checks](#testing) and add additional tests as necessary.
1. Open a pull request.
1. Ensure CI checks pass.

## Testing

Please ensure that all tests pass before submitting a pull request.  Please add
tests for any new functionality or bug fixes.

Additionally, this project leverages [pre-commit](#pre-commit) to perform linting
and formatting.  See instructions below for installing and running
[pre-commit](#pre-commit) as part of your development workflow.

### Unit Tests

This project uses the `testing` package from the Go standard library.  To run all
tests, execute the following command from within the project directory:

```bash
go test -v ./...
```

To run a specific test function, execute the following command from within the
project directory:

```bash
go test -v -run TestFunctionName ./...
```

### Pre-commit

This project leverages [pre-commit][1] to run tests and checks with every commit
and is verified again at the time of pull request.  After installing [pre-commit][1]
run the following command from within the project directory:

```bash
pre-commit install
```

#### Prerequisites

The `pre-commit` hooks require additional local dependencies to be installed.
These are needed to run checks locally, e.g., `golangci-lint`.

##### Go

We're using `golangci-lint` for linting Go code with a collection of common linters,
including `gofmt`, `govet`, `gosimple`, and `errcheck`.  Install `golangci-lint`:

```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.2
```

If the `GOPATH` is not part of your `PATH`, you may need to add it:

```bash
export PATH=$(go env GOPATH)/bin:$PATH
```

Some predefined linters have been setup in `.golangci.yaml` already and can be
executed with the following command:

```bash
golangci-lint run
```

##### Ruby

The `markdownlint` pre-commit hook requires `ruby` of version 2.7 or later.

[embedmd]:# (.pre-commit-config.yaml yaml /  - repo: https:\/\/github.com\/jumanjihouse\/pre-commit-hooks/ /id: markdownlint/)

```yaml
  - repo: https://github.com/jumanjihouse/pre-commit-hooks
    rev: 3.0.0
    hooks:
      - id: markdownlint
```

It is highly recommended to use [rbenv][2] to setup Ruby locally.  When deciding
on a version of Ruby, it is recommended to use the latest stable version from this
list:

```bash
# list latest stable versions:
rbenv install -l

# install a Ruby version (3.4.1 as an example):
rbenv install 3.4.1
```

> [!TIP]
> Don't forget to set the path to the Ruby version you installed:
> ```bash
> rvenv global 3.4.1
> eval "$(rbenv init -)"
> ```

A custom markdownlint rule has been written to work in conjunction with `embedmd`
and can be added to to any repo to likewise exclude line length checks for lines
prefixed by "[embedmd]".  See [.mdlrc](.mdlrc) for the referenced rule and configuration.

[embedmd]:# (.mdlrc yaml)

```yaml
rulesets ['ci/custom-line-length.rb']
style 'ci/markdown-style.rb'
```

#### Running pre-commit Manually

After installing `pre-commit` with `pre-commit install` it will automatically run
on commits.  At any time you can run the pre-commit checks manually, prior to
committing, with the following command:

```bash
pre-commit run --all-files
```

Alternatively, you can run the checks individually by specifying their hook `id`:

```bash
pre-commit run go-fmt --all-files
```

Individual files can be specified as well:

```bash
pre-commit run markdownlint --files README.md sample/docs.md
```

For additional information for running `pre-commit` run:

```bash
pre-commit run --help
```

## Opening PRs

When opening a pull request, please fill out the checklist supplied the template.
This will help others properly categorize and review your pull request.

### PR title

This project leverages squash merging and [conventional commits][3].  The following
conventional commit labels are accepted:

- feat
- fix
- docs
- test
- ci
- refactor
- perf
- chore
- revert

Example PR titles:

- “feat: add ability to authenticate to private repos”
- "fix: addresses Issue #1234 - check for closing parentheses in parser"

<!-- links -->
[1]: https://pre-commit.com/#install
[2]: https://github.com/rbenv/rbenv
[3]: https://www.conventionalcommits.org/en/v1.0.0/
