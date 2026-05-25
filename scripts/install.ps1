# install.ps1 - Windows installer for mju-dataset
# Usage: irm <install_script_url> | iex
#
# This script does NOT contain the labeling server address.
# It only downloads the CLI binary from GitHub Releases.
#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Repo    = "mjudcd-ct-r-d-labeling/labeling_download_cli"
$Binary  = "mju-dataset"
$Asset   = "mju-dataset-windows-amd64.exe"
$ChecksumFile = "checksums-windows-amd64.txt"

# ── Install directory ─────────────────────────────────────────────────────────
$InstallDir = Join-Path $env:LOCALAPPDATA "mju-dataset"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# ── Resolve latest release tag ────────────────────────────────────────────────
Write-Host "Fetching latest release..."
$ApiUrl  = "https://api.github.com/repos/$Repo/releases/latest"
$Release = Invoke-RestMethod -Uri $ApiUrl -Headers @{ 'User-Agent' = 'mju-dataset-installer' }
$Tag     = $Release.tag_name
if (-not $Tag) {
    Write-Error "Failed to determine the latest release."
    exit 1
}

Write-Host "Installing ${Binary} ${Tag} (windows/amd64)..."
$BaseUrl = "https://github.com/$Repo/releases/download/$Tag"

# ── Download to a temp directory ──────────────────────────────────────────────
$TmpDir = Join-Path $env:TEMP ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null

try {
    $BinaryPath   = Join-Path $TmpDir $Asset
    $ChecksumPath = Join-Path $TmpDir $ChecksumFile

    Write-Host "Downloading binary..."
    Invoke-WebRequest -Uri "$BaseUrl/$Asset"        -OutFile $BinaryPath   -UseBasicParsing
    Write-Host "Downloading checksum..."
    Invoke-WebRequest -Uri "$BaseUrl/$ChecksumFile" -OutFile $ChecksumPath -UseBasicParsing

    # ── Verify SHA-256 ────────────────────────────────────────────────────────
    Write-Host "Verifying checksum..."
    $ExpectedLine = Get-Content $ChecksumPath | Where-Object { $_ -match $Asset }
    if (-not $ExpectedLine) {
        Write-Error "Checksum entry for $Asset not found in $ChecksumFile."
        exit 1
    }
    $Expected = ($ExpectedLine -split '\s+')[0].ToLower()
    $Actual   = (Get-FileHash -Algorithm SHA256 -Path $BinaryPath).Hash.ToLower()
    if ($Actual -ne $Expected) {
        Write-Error "Checksum mismatch. The download may be corrupt or tampered.`nExpected: $Expected`nActual:   $Actual"
        exit 1
    }

    # ── Install ───────────────────────────────────────────────────────────────
    $Dest = Join-Path $InstallDir "${Binary}.exe"
    Copy-Item -Path $BinaryPath -Destination $Dest -Force

} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}

# ── PATH registration ─────────────────────────────────────────────────────────
$UserPath = [System.Environment]::GetEnvironmentVariable('Path', 'User')
if ($UserPath -notlike "*$InstallDir*") {
    [System.Environment]::SetEnvironmentVariable(
        'Path',
        "$UserPath;$InstallDir",
        'User'
    )
    Write-Host ""
    Write-Host "Added $InstallDir to your user PATH."
    Write-Host "Please restart your terminal for the change to take effect."
}

Write-Host ""
Write-Host "${Binary} installed successfully to $InstallDir\${Binary}.exe"
& (Join-Path $InstallDir "${Binary}.exe") --version
