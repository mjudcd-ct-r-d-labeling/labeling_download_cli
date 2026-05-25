# install.ps1 - Windows installer for mju-dataset
# Usage: irm <install_script_url> | iex
#
# This script does NOT contain the labeling server address.
# It fetches the latest release from the distribution API and downloads
# the binary via a presigned URL.
#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$ApiBase     = "https://mjudcd-grac-api.newlearn.ai.kr"
$Binary      = "mju-dataset"
$OsBuildType = "windows-amd64"

# ── Install directory ─────────────────────────────────────────────────────────
$InstallDir = Join-Path $env:LOCALAPPDATA "mju-dataset"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# ── Fetch latest version info ─────────────────────────────────────────────────
Write-Host "Fetching latest release for ${OsBuildType}..."
$LatestUrl = "${ApiBase}/cli-releases/latest?os_build_type=${OsBuildType}&current_version=0.0.0"
$Latest    = Invoke-RestMethod -Uri $LatestUrl -Headers @{ 'User-Agent' = 'mju-dataset-installer' }

$Version     = $Latest.version
$DownloadUrl = $Latest.download_url
$Sha256      = $Latest.sha256

if (-not $Version -or -not $DownloadUrl) {
    Write-Error "Failed to fetch latest release information."
    exit 1
}

Write-Host "Installing ${Binary} v${Version} (${OsBuildType})..."

# ── Download to a temp directory ──────────────────────────────────────────────
$TmpDir = Join-Path $env:TEMP ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null

try {
    $BinaryPath = Join-Path $TmpDir "${Binary}.exe"

    Write-Host "Downloading binary..."
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $BinaryPath -UseBasicParsing

    # ── Verify SHA-256 ────────────────────────────────────────────────────────
    if ($Sha256) {
        Write-Host "Verifying checksum..."
        $Actual = (Get-FileHash -Algorithm SHA256 $BinaryPath).Hash.ToLower()
        if ($Actual -ne $Sha256.ToLower()) {
            Write-Error "Checksum verification failed.`nExpected: $Sha256`nActual:   $Actual"
            exit 1
        }
        Write-Host "Checksum OK"
    }

    # ── Install ───────────────────────────────────────────────────────────────
    $Dest = Join-Path $InstallDir "${Binary}.exe"
    Copy-Item $BinaryPath $Dest -Force

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
Write-Host "${Binary} v${Version} installed to $InstallDir\${Binary}.exe"
& (Join-Path $InstallDir "${Binary}.exe") --version
