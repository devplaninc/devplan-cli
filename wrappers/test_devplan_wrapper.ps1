# Test script for devplan.ps1 wrapper
# Tests argument forwarding, exec parsing, exit code handling, and cleanup

$ErrorActionPreference = "Stop"

Write-Host "Testing devplan.ps1 wrapper..."

# Setup: Create temp directory for test artifacts
$TestDir = Join-Path $env:TEMP "devplan_wrapper_test_$([guid]::NewGuid().ToString('N'))"
New-Item -ItemType Directory -Path $TestDir -Force | Out-Null

$PassCount = 0
$FailCount = 0

function Pass {
    param([string]$Message)
    Write-Host ([char]0x2713 + " $Message") -ForegroundColor Green
    $script:PassCount++
}

function Fail {
    param([string]$Message)
    Write-Host ([char]0x2717 + " $Message") -ForegroundColor Red
    $script:FailCount++
}

try {
    # Copy the wrapper script to test directory
    $ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
    Copy-Item (Join-Path $ScriptDir "devplan.ps1") (Join-Path $TestDir "devplan.ps1")

    # Test 1: Argument forwarding with spaces and special characters
    Write-Host "`nTest 1: Argument forwarding"
    $mockScript = @'
$argsOutput = @()
$instFile = $null
foreach ($arg in $args) {
    if ($arg -match '^--instructions-file=(.+)$') {
        $instFile = $matches[1]
    }
    $argsOutput += "arg: $arg"
}
$argsOutput | ForEach-Object { Write-Host $_ }
exit 0
'@
    Set-Content -Path (Join-Path $TestDir "devplan.exe") -Value $mockScript

    # Rename .exe to .ps1 for testing (PowerShell mock)
    $mockBin = Join-Path $TestDir "devplan.ps1"
    # We need to modify the wrapper to call .ps1 instead of .exe for testing
    $wrapperContent = Get-Content (Join-Path $TestDir "devplan.ps1") -Raw
    $testWrapper = $wrapperContent -replace 'devplan\.exe', 'devplan_mock.ps1'
    Set-Content -Path (Join-Path $TestDir "devplan_test.ps1") -Value $testWrapper

    # Create mock as PS1
    Set-Content -Path (Join-Path $TestDir "devplan_mock.ps1") -Value $mockScript

    $output = & powershell -NoProfile -File (Join-Path $TestDir "devplan_test.ps1") "arg with spaces" "special!@#" "--flag=value" 2>&1
    $outputStr = $output -join "`n"

    if ($outputStr -match 'arg: arg with spaces' -and
        $outputStr -match 'arg: special!@#' -and
        $outputStr -match 'arg: --flag=value' -and
        $outputStr -match 'arg: --instructions-file=') {
        Pass "Arguments forwarded correctly with spaces and special chars"
    } else {
        Fail "Argument forwarding failed"
        Write-Host "Output: $outputStr"
    }

    # Test 2: Valid exec line - returns exec command exit code
    Write-Host "`nTest 2: Valid exec line execution"
    $mockScript = @'
foreach ($arg in $args) {
    if ($arg -match '^--instructions-file=(.+)$') {
        Set-Content -Path $matches[1] -Value 'exec: "exit 42"'
    }
}
exit 0
'@
    Set-Content -Path (Join-Path $TestDir "devplan_mock.ps1") -Value $mockScript

    & powershell -NoProfile -File (Join-Path $TestDir "devplan_test.ps1") "test-arg"
    $exitCode = $LASTEXITCODE

    if ($exitCode -eq 42) {
        Pass "Exec command exit code (42) returned correctly"
    } else {
        Fail "Expected exit code 42, got $exitCode"
    }

    # Test 3: Missing instructions file - returns CLI exit code
    Write-Host "`nTest 3: Missing instructions file"
    $mockScript = @'
foreach ($arg in $args) {
    if ($arg -match '^--instructions-file=(.+)$') {
        Remove-Item $matches[1] -Force -ErrorAction SilentlyContinue
    }
}
exit 5
'@
    Set-Content -Path (Join-Path $TestDir "devplan_mock.ps1") -Value $mockScript

    & powershell -NoProfile -File (Join-Path $TestDir "devplan_test.ps1") "test-arg"
    $exitCode = $LASTEXITCODE

    if ($exitCode -eq 5) {
        Pass "CLI exit code (5) returned when file missing"
    } else {
        Fail "Expected exit code 5, got $exitCode"
    }

    # Test 4: Empty instructions file - returns CLI exit code
    Write-Host "`nTest 4: Empty instructions file"
    $mockScript = @'
foreach ($arg in $args) {
    if ($arg -match '^--instructions-file=(.+)$') {
        Set-Content -Path $matches[1] -Value ""
    }
}
exit 3
'@
    Set-Content -Path (Join-Path $TestDir "devplan_mock.ps1") -Value $mockScript

    & powershell -NoProfile -File (Join-Path $TestDir "devplan_test.ps1") "test-arg"
    $exitCode = $LASTEXITCODE

    if ($exitCode -eq 3) {
        Pass "CLI exit code (3) returned for empty file"
    } else {
        Fail "Expected exit code 3, got $exitCode"
    }

    # Test 5: Malformed exec line (no quotes) - returns CLI exit code
    Write-Host "`nTest 5: Malformed exec line"
    $mockScript = @'
foreach ($arg in $args) {
    if ($arg -match '^--instructions-file=(.+)$') {
        Set-Content -Path $matches[1] -Value "exec: exit 99"
    }
}
exit 4
'@
    Set-Content -Path (Join-Path $TestDir "devplan_mock.ps1") -Value $mockScript

    & powershell -NoProfile -File (Join-Path $TestDir "devplan_test.ps1") "test-arg"
    $exitCode = $LASTEXITCODE

    if ($exitCode -eq 4) {
        Pass "CLI exit code (4) returned for malformed exec line"
    } else {
        Fail "Expected exit code 4, got $exitCode"
    }

    # Test 6: CLI returns non-zero with no exec - returns CLI exit code
    Write-Host "`nTest 6: CLI non-zero exit with no exec"
    $mockScript = @'
foreach ($arg in $args) {
    if ($arg -match '^--instructions-file=(.+)$') {
        Set-Content -Path $matches[1] -Value "# no exec here"
    }
}
exit 7
'@
    Set-Content -Path (Join-Path $TestDir "devplan_mock.ps1") -Value $mockScript

    & powershell -NoProfile -File (Join-Path $TestDir "devplan_test.ps1") "test-arg"
    $exitCode = $LASTEXITCODE

    if ($exitCode -eq 7) {
        Pass "CLI exit code (7) returned when no exec present"
    } else {
        Fail "Expected exit code 7, got $exitCode"
    }

    # Test 7: Exec command with spaces
    Write-Host "`nTest 7: Exec command with spaces"
    $mockScript = @'
foreach ($arg in $args) {
    if ($arg -match '^--instructions-file=(.+)$') {
        Set-Content -Path $matches[1] -Value 'exec: "Write-Host ''hello world''; exit 0"'
    }
}
exit 0
'@
    Set-Content -Path (Join-Path $TestDir "devplan_mock.ps1") -Value $mockScript

    $output = & powershell -NoProfile -File (Join-Path $TestDir "devplan_test.ps1") "test-arg" 2>&1
    $exitCode = $LASTEXITCODE
    $outputStr = $output -join "`n"

    if ($exitCode -eq 0 -and $outputStr -match 'hello world') {
        Pass "Exec command with spaces executed correctly"
    } else {
        Fail "Exec command with spaces failed (exit: $exitCode, output: $outputStr)"
    }

    # Test 8: Whitespace handling in exec line
    Write-Host "`nTest 8: Whitespace handling"
    $mockScript = @'
foreach ($arg in $args) {
    if ($arg -match '^--instructions-file=(.+)$') {
        Set-Content -Path $matches[1] -Value '  exec:   "exit 13"  '
    }
}
exit 0
'@
    Set-Content -Path (Join-Path $TestDir "devplan_mock.ps1") -Value $mockScript

    & powershell -NoProfile -File (Join-Path $TestDir "devplan_test.ps1") "test-arg"
    $exitCode = $LASTEXITCODE

    if ($exitCode -eq 13) {
        Pass "Whitespace in exec line handled correctly"
    } else {
        Fail "Expected exit code 13, got $exitCode"
    }

    # Summary
    Write-Host "`n========================================="
    Write-Host "Test Results: $PassCount passed, $FailCount failed"
    Write-Host "========================================="

    if ($FailCount -gt 0) {
        exit 1
    }

    Write-Host "All tests passed!"
}
finally {
    # Cleanup
    if (Test-Path $TestDir) {
        Remove-Item $TestDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
