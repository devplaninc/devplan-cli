# Devplan CLI Installer for Windows (PowerShell)

$ErrorActionPreference = "Stop"

# Detect architecture
$arch = $env:PROCESSOR_ARCHITECTURE
switch ($arch) {
    "AMD64" { $arch = "amd64" }
    "ARM64" { $arch = "arm64" }
    default {
        Write-Error "Unsupported architecture: $arch"
        exit 1
    }
}

# Set install directories
$InstallDir = "$env:USERPROFILE\bin"
$BinaryName = "devplan.exe"
$Version = $env:DEVPLAN_INSTALL_VERSION
if (-not $Version) { $Version = "latest" }

# Fetch latest version if needed
if ($Version -eq "latest") {
    $configUrl = "https://devplan-cli.sfo3.digitaloceanspaces.com/releases/version.json"
    Write-Host "Fetching latest version from $configUrl..."
    $json = Invoke-RestMethod -Uri $configUrl
    $Version = $json.productionVersion
    if (-not $Version) {
        Write-Error "Failed to extract production version"
        exit 1
    }
    Write-Host "Latest version is $Version"
}

# Build download URL
$downloadUrl = "https://devplan-cli.sfo3.digitaloceanspaces.com/releases/versions/$Version/devplan-windows-$arch.zip"
$tempZip = "$env:TEMP\devplan.zip"
$tempExtract = "$env:TEMP\devplan_extracted"

Write-Host "Downloading Devplan CLI from $downloadUrl"
Invoke-WebRequest -Uri $downloadUrl -OutFile $tempZip

# Extract
Expand-Archive -Path $tempZip -DestinationPath $tempExtract -Force
$sourceExe = Join-Path $tempExtract "devplan-windows-$arch.exe"
$targetExe = Join-Path $InstallDir $BinaryName

# Create install dir if needed
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Move executable
Move-Item -Force -Path $sourceExe -Destination $targetExe

# Clean up
Remove-Item $tempZip -Force
Remove-Item $tempExtract -Recurse -Force

Write-Host "`n✅ Devplan CLI installed to: $targetExe"

# Check if it's in PATH
$inPath = ($env:PATH -split ';') -contains $InstallDir
if (-not $inPath) {
    Write-Warning "`n$InstallDir is not in your PATH."
    Write-Host "You can add it using:"
    Write-Host "[System.Environment]::SetEnvironmentVariable('Path', \$env:Path + ';$InstallDir', 'User')"
    Write-Host "Then restart your terminal."
} else {
    Write-Host "`nYou can now run: devplan --help"
}
