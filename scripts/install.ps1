# install.ps1 - Windows installer for mju-dataset
# Usage: irm <install_script_url> | iex
#
# This script does NOT contain the labeling server address.
# It fetches the latest release from the distribution API and downloads
# the binary via a presigned URL.
#Requires -Version 5.1
param(
    # A specific version can be passed as an argument:
    # & ([scriptblock]::Create((irm '<url>/install.ps1'))) -Version '2026.05.26.42'
    [string]$Version = ""
)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$ApiBase     = "https://mjudcd-grac-api.newlearn.ai.kr"
$GhRepo      = "mjudcd-ct-r-d-labeling/labeling_download_cli"
$Binary      = "mju-dataset"
$OsBuildType = "windows-amd64"

# ── Install directory ─────────────────────────────────────────────────────────
$InstallDir = Join-Path $env:LOCALAPPDATA "mju-dataset"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# ── Fetch latest version from Git tags (인자가 없을 때만) ────────────────────────────────
if (-not $Version) {
    Write-Host "Fetching latest release for ${OsBuildType}..."
    $TagsUrl = "https://api.github.com/repos/$GhRepo/tags?per_page=1"
    $Tags    = Invoke-RestMethod -Uri $TagsUrl `
        -Headers @{ 'Accept' = 'application/vnd.github+json'; 'User-Agent' = 'mju-dataset-installer' }
    $Version = $Tags[0].name
    if (-not $Version) {
        Write-Error "Failed to determine latest version from GitHub tags."
        exit 1
    }
}

# ── Fetch download URL from release server ────────────────────────────────────────────
$DlUrl  = "${ApiBase}/cli-releases/download/${OsBuildType}/${Version}"
$DlInfo = Invoke-RestMethod -Uri $DlUrl `
    -Headers @{ 'User-Agent' = 'mju-dataset-installer' }

$DownloadUrl = $DlInfo.download_url
$Sha256      = $DlInfo.sha256

if (-not $DownloadUrl) {
    Write-Error "Failed to parse download URL for ${OsBuildType} v${Version}."
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
