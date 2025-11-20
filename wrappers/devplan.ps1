# Devplan CLI Wrapper Script (PowerShell)
# This script wraps the devplan binary to support follow-up instructions execution.
# It forwards all arguments to the CLI with an additional --instructions-file parameter,
# then executes any command specified in the instructions file.

$ErrorActionPreference = "Stop"

# Resolve the directory where this script is located
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$DevplanBin = Join-Path $ScriptDir "devplan.exe"

# Check if binary exists
if (-not (Test-Path $DevplanBin)) {
    Write-Error "Error: devplan binary not found at $DevplanBin"
    exit 1
}

# Create temporary instructions file
$InstructionsFile = [System.IO.Path]::GetTempFileName()
if (-not $InstructionsFile -or -not (Test-Path $InstructionsFile)) {
    Write-Error "Error: Failed to create temporary instructions file"
    exit 1
}

try {
    # Run the devplan CLI with all arguments plus the instructions file
    $cliArgs = @($args) + @("--instructions-file=$InstructionsFile")
    & $DevplanBin @cliArgs
    $CliExitCode = $LASTEXITCODE

    # If instructions file doesn't exist or is empty, exit with CLI exit code
    if (-not (Test-Path $InstructionsFile)) {
        exit $CliExitCode
    }

    $fileContent = Get-Content $InstructionsFile -Raw -ErrorAction SilentlyContinue
    if (-not $fileContent -or $fileContent.Trim() -eq "") {
        exit $CliExitCode
    }

    # Parse the first non-empty line for exec: "command" pattern
    $execCmd = $null
    $lines = $fileContent -split "`r?`n"

    foreach ($line in $lines) {
        # Skip empty lines
        if ($line.Trim() -eq "") {
            continue
        }

        # Check if line matches exec: "..." pattern
        if ($line -match '^\s*exec:\s*"([^"]*)"') {
            $execCmd = $matches[1]
            break
        }

        # Only check first non-empty line
        break
    }

    # If no valid exec command found, exit with CLI exit code
    if (-not $execCmd) {
        exit $CliExitCode
    }

    # Execute the extracted command
    powershell -NoProfile -Command $execCmd
    $ExecExitCode = $LASTEXITCODE

    exit $ExecExitCode
}
finally {
    # Cleanup: remove temporary file
    if (Test-Path $InstructionsFile) {
        Remove-Item $InstructionsFile -Force -ErrorAction SilentlyContinue
    }
}
