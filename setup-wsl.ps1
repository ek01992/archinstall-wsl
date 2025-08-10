# Requires Windows 10 2004+ or Windows 11 with WSL2 enabled

[CmdletBinding()]
param(
    [switch]$Force,
    [string]$DistroName = 'archlinux',
    [string]$DefaultUser = 'erik',
    [ValidateSet('static','resolved','wsl')]
    [string]$DnsMode = 'static',
    [int]$SwapGB = 4
)

$ErrorActionPreference = 'Stop'

# -------------------------------
# Configuration
# -------------------------------
$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$bootstrapLocal = Join-Path $repoRoot "bootstrap.sh"
$snapshotFile = Join-Path $env:USERPROFILE "arch.tar.bak"
$wslConfigPath = Join-Path $env:USERPROFILE ".wslconfig"

$WSLMemory = if ($env:WSL_MEMORY) { $env:WSL_MEMORY } else { "8GB" }
$WSLCPUs   = if ($env:WSL_CPUS)   { $env:WSL_CPUS }   else { "4" }

# Execution policy softening for CurrentUser
try {
  $scope = 'CurrentUser'
  $policy = Get-ExecutionPolicy -Scope $scope -ErrorAction SilentlyContinue
  if ($policy -ne 'RemoteSigned' -and $policy -ne 'Unrestricted') {
    Set-ExecutionPolicy -Scope $scope -ExecutionPolicy RemoteSigned -Force -ErrorAction Stop
  }
} catch {
  Write-Warning "Failed to set ExecutionPolicy; continuing. Error: $($_.Exception.Message)"
}

function Write-Preflight {
  Write-Host "[*] Preflight summary:"
  Write-Host "    Distro:           $DistroName"
  Write-Host "    Default user:     $DefaultUser"
  Write-Host "    CPUs:             $WSLCPUs"
  Write-Host "    Memory:           $WSLMemory"
  Write-Host "    Swap:             ${SwapGB}GB"
  Write-Host "    DNS mode:         $DnsMode"
  Write-Host "    Snapshot target:  $snapshotFile"
}

function Set-WSLDefault {
  $null = & wsl.exe --status 2>$null
  Write-Host "[*] WSL is installed and available."
  if ($LASTEXITCODE -ne 0) {
    throw "WSL is not installed or not available. Install WSL and WSL2 first."
  }
  # Try to set default version 2 (best-effort)
  try { $null = & wsl.exe --set-default-version 2 } catch { }
  Write-Host "[*] WSL default version set to 2."
}

function Write-WslConfigSafe {
  # Back up existing file if present and not ours
  $content = @"
[wsl2]
memory=$WSLMemory
processors=$WSLCPUs
swap=$(${SwapGB})GB
localhostForwarding=true
"@
  if (Test-Path $wslConfigPath) {
    $existing = Get-Content -Raw $wslConfigPath
    if ($existing -ne $content) {
      $backup = "$wslConfigPath.bak"
      Copy-Item $wslConfigPath $backup -Force
      Write-Host "[i] Backed up existing .wslconfig to $backup"
    }
  }
  $content | Set-Content -NoNewline -Encoding ASCII $wslConfigPath
  Write-Host "[*] Wrote $wslConfigPath (memory=$WSLMemory, processors=$WSLCPUs, swap=${SwapGB}GB)"
}

function Get-DistroList {
  & wsl.exe --list --quiet | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne '' }
}

function Wait-ForDistro {
  param([string]$Name, [int]$TimeoutSec = 60)
  $sw = [Diagnostics.Stopwatch]::StartNew()
  while ($sw.Elapsed.TotalSeconds -lt $TimeoutSec) {
    $list = Get-DistroList
    if ($list -contains $Name) { return $true }
    Start-Sleep -Seconds 2
  }
  return $false
}

function Invoke-WSLChecked {
  param([string]$Arguments, [string]$ErrorContext)
  & wsl.exe $Arguments
  if ($LASTEXITCODE -ne 0) {
    throw "$ErrorContext failed with exit code $LASTEXITCODE"
  }
}

Write-Preflight
Set-WSLDefault
Write-WslConfigSafe

Write-Host "[*] Applying .wslconfig (wsl --shutdown)..."
& wsl.exe --shutdown
Start-Sleep -Milliseconds 1500

Write-Host "[*] Checking WSL installation..."
& wsl.exe --status | Out-Null

Write-Host "[*] Preparing distro '$DistroName'..."
$installed = Get-DistroList
if ($installed -contains $DistroName) {
  if (-not $Force) {
    $resp = Read-Host "Distro '$DistroName' exists. Unregister it and reinstall? (y/N)"
    if ($resp -notin @("y","Y")) {
      throw "Aborted by user."
    }
  }
  Write-Host "[*] Unregistering '$DistroName'..."
  & wsl.exe --terminate $DistroName | Out-Null
  & wsl.exe --unregister $DistroName
}

Write-Host "[*] Installing '$DistroName'..."
try {
  & wsl.exe --install --no-launch -d $DistroName
} catch {
  Write-Warning "wsl.exe --install failed. If the Microsoft Store Arch is unavailable, consider importing an Arch rootfs manually."
  throw
}

if (-not (Wait-ForDistro -Name $DistroName -TimeoutSec 120)) {
  throw "Distro '$DistroName' did not appear after install."
}

# Prepare bootstrap path with normalized LF endings to avoid /usr/bin/env bash\r errors
$tempBootstrap = Join-Path $env:TEMP "bootstrap-arch-wsl.sh"
# Normalize CRLF -> LF and write UTF-8 without BOM
$text = Get-Content -Raw -Encoding UTF8 $bootstrapLocal
$text = $text -replace "`r`n","`n"
$text = $text -replace "`r","`n"
Set-Content -Path $tempBootstrap -Value $text -NoNewline -Encoding utf8NoBOM
Write-Host "[*] Prepared $tempBootstrap (LF line endings)."

# Guard UNC paths (\\server\share\...) which WSL /mnt mapping cannot construct
if ($tempBootstrap.StartsWith("\\") -or $repoRoot.StartsWith("\\")) {
  throw "Repository or temp path is a UNC path. Please use a local drive path for correct /mnt mapping."
}

$mntPath = "/mnt/" + ($tempBootstrap.Substring(0,1).ToLower()) + ($tempBootstrap.Substring(2) -replace "\\","/")
$mntRepoRoot = "/mnt/" + ($repoRoot.Substring(0,1).ToLower()) + ($repoRoot.Substring(2) -replace "\\","/")

Write-Host "[*] What this script will do:"
Write-Host "    - Phase 1: base system, keyring, packages, user, DNS($DnsMode), dotfiles"
Write-Host "    - Restart WSL to enable systemd & default user"
Write-Host "    - Phase 2: enable services and finalize toolchains"
Write-Host ""

Write-Host "[*] Running Phase 1 inside WSL as root..."
Invoke-WSLChecked "-d $DistroName -u root -- bash -lc ""chmod +x '$mntPath' && DEFAULT_USER='$DefaultUser' WSL_MEMORY='$WSLMemory' WSL_CPUS='$WSLCPUs' REPO_ROOT_MNT='$mntRepoRoot' DNS_MODE='$DnsMode' '$mntPath' phase1""" "Phase 1"

Write-Host "[*] Shutting down WSL to activate systemd and default user..."
& wsl.exe --shutdown

Write-Host "[*] Running Phase 2 inside WSL..."
$phase2Cmd = "DEFAULT_USER='$DefaultUser' DNS_MODE='$DnsMode' '$mntPath' phase2"
Invoke-WSLChecked "-d $DistroName -u root -- bash -lc ""$phase2Cmd""" "Phase 2"

Write-Host "[*] Terminating distro before export..."
& wsl.exe --terminate $DistroName

Write-Host "[*] Exporting clean snapshot (warnings about sockets are harmless)..."
& wsl.exe --export $DistroName $snapshotFile
Write-Host "[+] Snapshot saved to: $snapshotFile"

Write-Host ""
Write-Host "Reset steps:"
Write-Host "  wsl --unregister $DistroName"
Write-Host "  wsl --import $DistroName C:\WSL\Arch $snapshotFile --version 2"
Write-Host ""
Write-Host "Done."