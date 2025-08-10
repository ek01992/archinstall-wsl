# setup-wsl.ps1
# Requires Windows 10 2004+ or Windows 11 with WSL2 enabled

[CmdletBinding(SupportsShouldProcess=$true, ConfirmImpact='Medium')]
param(
    [switch]$Force,
    [string]$DistroName = 'archlinux',
    [string]$DefaultUser = 'erik',
    [ValidateSet('static','resolved','wsl')]
    [string]$DnsMode = 'static',
    [int]$SwapGB = 4
)

Set-StrictMode -Version Latest
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

function Write-Status {
  param([string]$Message, [ConsoleColor]$Color = [ConsoleColor]::Gray)
  Write-Host $Message -ForegroundColor $Color
}

function Ensure-ExecutionPolicy {
  try {
    $scope = 'CurrentUser'
    $policy = Get-ExecutionPolicy -Scope $scope -ErrorAction SilentlyContinue
    if ($policy -ne 'RemoteSigned' -and $policy -ne 'Unrestricted') {
      if ($PSCmdlet.ShouldProcess("Set-ExecutionPolicy ($scope)", "RemoteSigned")) {
        Set-ExecutionPolicy -Scope $scope -ExecutionPolicy RemoteSigned -Force -ErrorAction Stop
      }
    }
  } catch {
    Write-Warning "Failed to set ExecutionPolicy; continuing. Error: $($_.Exception.Message)"
  }
}

function Test-WslAvailable {
  & wsl.exe --status *> $null
  return ($LASTEXITCODE -eq 0)
}

function Set-WslDefaultVersion2 {
  if ($PSCmdlet.ShouldProcess("WSL default version", "Set to 2")) {
    try { $null = & wsl.exe --set-default-version 2 } catch { }
    Write-Status "[*] WSL default version set to 2." -Color Cyan
  }
}

function Get-WslDistributions {
  & wsl.exe --list --quiet | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne '' }
}

function Test-WslDistributionExists {
  param([string]$Name)
  $d = Get-WslDistributions
  return ($d -contains $Name)
}

function Wait-WslDistribution {
  param([string]$Name, [int]$TimeoutSec = 60)
  $sw = [Diagnostics.Stopwatch]::StartNew()
  while ($sw.Elapsed.TotalSeconds -lt $TimeoutSec) {
    if (Test-WslDistributionExists -Name $Name) { return $true }
    Start-Sleep -Seconds 2
  }
  return $false
}

function Set-WslConfigFile {
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
      if ($PSCmdlet.ShouldProcess(".wslconfig", "Backup to $backup")) {
        Copy-Item $wslConfigPath $backup -Force
        Write-Status "[i] Backed up existing .wslconfig to $backup" -Color DarkGray
      }
    }
  }
  if ($PSCmdlet.ShouldProcess(".wslconfig", "Write memory=$WSLMemory, processors=$WSLCPUs, swap=${SwapGB}GB")) {
    $content | Set-Content -NoNewline -Encoding ASCII $wslConfigPath
    Write-Status "[*] Wrote $wslConfigPath (memory=$WSLMemory, processors=$WSLCPUs, swap=${SwapGB}GB)" -Color Cyan
  }
}

function ConvertTo-WslPath {
  param([Parameter(Mandatory)][string]$WindowsPath)
  if ($WindowsPath.StartsWith("\\") ) {
    throw "UNC path not supported by WSL /mnt mapping: '$WindowsPath'"
  }
  $drive = $WindowsPath.Substring(0,1).ToLower()
  $rest = $WindowsPath.Substring(2).Replace("\","/")
  return "/mnt/$drive$rest"
}

function Prepare-BootstrapScript {
  # Normalize LF endings to avoid /usr/bin/env bash\r errors
  $tempBootstrap = Join-Path $env:TEMP "bootstrap-arch-wsl.sh"
  $text = Get-Content -Raw -Encoding UTF8 $bootstrapLocal
  $text = $text -replace "`r`n","`n"
  $text = $text -replace "`r","`n"
  Set-Content -Path $tempBootstrap -Value $text -NoNewline -Encoding utf8NoBOM
  Write-Status "[*] Prepared $tempBootstrap (LF line endings)." -Color Cyan
  return $tempBootstrap
}

function Invoke-WslChecked {
  param(
    [string[]]$ArgList,
    [string]$ErrorContext
  )
  & wsl.exe @ArgList
  if ($LASTEXITCODE -ne 0) {
    throw "$ErrorContext failed with exit code $LASTEXITCODE"
  }
}

function Install-WslDistribution {
  if (Test-WslDistributionExists -Name $DistroName) {
    if (-not $Force) {
      $resp = Read-Host "Distro '$DistroName' exists. Unregister it and reinstall? (y/N)"
      if ($resp -notin @("y","Y")) {
        throw "Aborted by user."
      }
    }
    if ($PSCmdlet.ShouldProcess("WSL distro '$DistroName'", "Unregister")) {
      Write-Status "[*] Unregistering '$DistroName'..." -Color Yellow
      & wsl.exe --terminate $DistroName | Out-Null
      & wsl.exe --unregister $DistroName
    }
  }
  if ($PSCmdlet.ShouldProcess("WSL distro '$DistroName'", "Install --no-launch")) {
    Write-Status "[*] Installing '$DistroName'..." -Color Cyan
    try {
      & wsl.exe --install --no-launch -d $DistroName
    } catch {
      Write-Warning "wsl.exe --install failed. If the Microsoft Store Arch is unavailable, consider importing an Arch rootfs manually."
      throw
    }
    if (-not (Wait-WslDistribution -Name $DistroName -TimeoutSec 120)) {
      throw "Distro '$DistroName' did not appear after install."
    }
  }
}

function Invoke-WslPhase {
  param(
    [Parameter(Mandatory)][ValidateSet('phase1','phase2')] [string]$Phase,
    [Parameter(Mandatory)][string]$MntBootstrap,
    [Parameter(Mandatory)][string]$MntRepoRoot
  )
  if ($Phase -eq 'phase1') {
    $cmd = "chmod +x '$MntBootstrap' && DEFAULT_USER='$DefaultUser' WSL_MEMORY='$WSLMemory' WSL_CPUS='$WSLCPUs' REPO_ROOT_MNT='$MntRepoRoot' DNS_MODE='$DnsMode' '$MntBootstrap' phase1"
    if ($PSCmdlet.ShouldProcess("WSL '$DistroName'", "Run Phase 1 as root")) {
      Invoke-WslChecked -ArgList @('-d', $DistroName, '-u', 'root', '--', 'bash', '-lc', $cmd) -ErrorContext "Phase 1"
    }
  } else {
    $cmd = "DEFAULT_USER='$DefaultUser' DNS_MODE='$DnsMode' '$MntBootstrap' phase2"
    if ($PSCmdlet.ShouldProcess("WSL '$DistroName'", "Run Phase 2")) {
      Invoke-WslChecked -ArgList @('-d', $DistroName, '-u', 'root', '--', 'bash', '-lc', $cmd) -ErrorContext "Phase 2"
    }
  }
}

function Export-WslSnapshot {
  if ($PSCmdlet.ShouldProcess("WSL '$DistroName'", "Terminate before export")) {
    & wsl.exe --terminate $DistroName
  }
  if ($PSCmdlet.ShouldProcess("WSL '$DistroName'", "Export to $snapshotFile")) {
    Write-Status "[*] Exporting clean snapshot (warnings about sockets are harmless)..." -Color Cyan
    & wsl.exe --export $DistroName $snapshotFile
    Write-Status "[+] Snapshot saved to: $snapshotFile" -Color Green
  }
}

function Show-Preflight {
  Write-Status "[*] Preflight summary:" -Color Cyan
  Write-Status ("    Distro:           {0}" -f $DistroName)
  Write-Status ("    Default user:     {0}" -f $DefaultUser)
  Write-Status ("    CPUs:             {0}" -f $WSLCPUs)
  Write-Status ("    Memory:           {0}" -f $WSLMemory)
  Write-Status ("    Swap:             {0}GB" -f $SwapGB)
  Write-Status ("    DNS mode:         {0}" -f $DnsMode)
  Write-Status ("    Snapshot target:  {0}" -f $snapshotFile)
}

function Show-Plan {
  Write-Status "[*] What this script will do:" -Color Cyan
  Write-Status "    - Phase 1: base system, keyring, packages, user, DNS($DnsMode), dotfiles"
  Write-Status "    - Restart WSL to enable systemd & default user"
  Write-Status "    - Phase 2: enable services and finalize toolchains"
  Write-Host ""
}

function Main {
  Show-Preflight
  Ensure-ExecutionPolicy

  if (-not (Test-WslAvailable)) {
    throw "WSL is not installed or not available. Install WSL and WSL2 first."
  }
  Write-Status "[*] WSL is installed and available." -Color Cyan
  Set-WslDefaultVersion2
  Set-WslConfigFile

  Write-Status "[*] Applying .wslconfig (wsl --shutdown)..." -Color Cyan
  if ($PSCmdlet.ShouldProcess("WSL", "Shutdown")) {
    & wsl.exe --shutdown
    Start-Sleep -Milliseconds 1500
  }

  Write-Status "[*] Checking WSL installation..." -Color DarkGray
  & wsl.exe --status | Out-Null

  Show-Plan

  # Prepare bootstrap script with LF endings and map to /mnt path
  $tempBootstrap = Prepare-BootstrapScript
  if ($tempBootstrap.StartsWith("\\") -or $repoRoot.StartsWith("\\")) {
    throw "Repository or temp path is a UNC path. Please use a local drive path for correct /mnt mapping."
  }
  $mntBootstrap = ConvertTo-WslPath -WindowsPath $tempBootstrap
  $mntRepoRoot  = ConvertTo-WslPath -WindowsPath $repoRoot

  Write-Status "[*] Preparing distro '$DistroName'..." -Color Cyan
  Install-WslDistribution

  Write-Status "[*] Running Phase 1 inside WSL as root..." -Color Cyan
  Invoke-WslPhase -Phase 'phase1' -MntBootstrap $mntBootstrap -MntRepoRoot $mntRepoRoot

  Write-Status "[*] Shutting down WSL to activate systemd and default user..." -Color Cyan
  if ($PSCmdlet.ShouldProcess("WSL", "Shutdown")) {
    & wsl.exe --shutdown
  }

  Write-Status "[*] Running Phase 2 inside WSL..." -Color Cyan
  Invoke-WslPhase -Phase 'phase2' -MntBootstrap $mntBootstrap -MntRepoRoot $mntRepoRoot

  Export-WslSnapshot

  Write-Host ""
  Write-Status "Reset steps:" -Color Cyan
  Write-Status "  wsl --unregister $DistroName"
  Write-Status "  wsl --import $DistroName C:\WSL\Arch $snapshotFile --version 2"
  Write-Host ""
  Write-Status "Done." -Color Green
}

Main