# Devplan CLI

A command-line interface for Devplan that helps automate development workflows.

## Version

You can check the current version of the CLI by running:

```bash
devplan version
```

## Self-Update

The CLI can update itself to the latest production version or to a specific version. Note that auto-update functionality is only available in binaries that have it enabled.

### Update to the latest production version

```bash
devplan update
```

### Update to a specific version

```bash
devplan update --to=1.2.3
```

### List available versions

```bash
devplan update --list
```

**Note for Homebrew users**: If you installed via Homebrew, the auto-update functionality is disabled. Please use Homebrew to update instead:

```bash
brew update && brew upgrade devplan
```

## Preferences

The CLI saves your preferences for company, project, and git protocol selections, and uses them as defaults on subsequent runs.

### Reset preferences

You can reset all saved preferences by running:

```bash
devplan prefs reset
```

This will clear the saved preferences for company, project, and git protocol from the local config file.

## Installation

### Direct Installation (Recommended for most users)

```bash
/bin/bash -c "$(curl -fsSL https://app.devplan.com/api/cli/install)"
```

This installs the binary with auto-update functionality enabled.

### Homebrew Installation

```bash
brew tap devplaninc/devplan
brew install devplan
```

The Homebrew version has auto-update functionality disabled. Use `brew upgrade devplan` to update.

### Manual Installation

Download the appropriate binary for your platform from the [GitHub Releases](https://github.com/devplaninc/devplan-cli/releases):

- For direct installation with auto-update: Use `devplan-<platform>-<arch>.tar.gz` or `devplan-<platform>-<arch>.zip`
- For package managers: Use `devplan-<platform>-<arch>-noautoupdate.tar.gz` or `devplan-<platform>-<arch>-noautoupdate.zip`

## Development Set up

### Private Github repo

We depend on github.com/devplaninc/webapp, which is a private repo. For that to work:

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

### Building with version information

The CLI uses build-time flags to embed version information. When building locally, you can use:

```bash
make build
```

This will set the version to "dev" for local development builds.

### Releasing a new version

1. Create and push a new tag with a semantic version (e.g., `v1.2.3`):
```bash
git tag v1.2.3
git push origin v1.2.3
```

2. This will trigger the GitHub Actions workflow to build and publish the release.

The release process automatically creates two sets of binaries for each supported platform:
- `devplan-<platform>-<arch>.*` - With auto-update functionality enabled
- `devplan-<platform>-<arch>-noautoupdate.*` - With auto-update functionality disabled (for package managers)

### Marking a version as production

After a version has been released and tested, you can mark it as the production version:

1. Go to the GitHub repository
2. Navigate to Actions → Mark Version as Production
3. Click "Run workflow"
4. Enter the version number (without the 'v' prefix, e.g., "1.2.3")
5. Click "Run workflow"

This will update the production version file in the Digital Ocean Space, and users will be able to update to this version using `devplan update`.
