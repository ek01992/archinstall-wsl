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
$snapshotFile = Join-Path $env:USERPROFILE "arch.tar.bak"

$WSLMemory = if ($env:WSL_MEMORY) { $env:WSL_MEMORY } else { "8GB" }
$WSLCPUs   = if ($env:WSL_CPUS)   { $env:WSL_CPUS }   else { "4" }

# Import module
Import-Module (Join-Path $repoRoot 'src/ps/ArchWSL.psm1') -Force

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

function Wait-ForDistro {
  param([string]$Name, [int]$TimeoutSec = 120)
  $sw = [Diagnostics.Stopwatch]::StartNew()
  while ($sw.Elapsed.TotalSeconds -lt $TimeoutSec) {
    $list = Get-ArchDistroList
    if ($list -contains $Name) { return $true }
    Start-Sleep -Seconds 2
  }
  return $false
}

Write-Preflight

# Ensure WSL defaults and write .wslconfig
try { $null = & wsl.exe --status 2>$null } catch { throw "WSL is not installed or not available." }
try { $null = & wsl.exe --set-default-version 2 } catch {}
Set-ArchWslConfig -Memory $WSLMemory -Processors $WSLCPUs -SwapGB $SwapGB

Write-Host "[*] Applying .wslconfig (wsl --shutdown)..."
& wsl.exe --shutdown
Start-Sleep -Milliseconds 1500

Write-Host "[*] Installing or resetting '$DistroName'..."
New-ArchDistro -DistroName $DistroName -Force:$Force

if (-not (Wait-ForDistro -Name $DistroName -TimeoutSec 120)) {
  throw "Distro '$DistroName' did not appear after install."
}

Write-Host "[*] What this script will do:"
Write-Host "    - Phase 1: base system, keyring, packages, user, DNS($DnsMode), dotfiles"
Write-Host "    - Restart WSL to enable systemd & default user"
Write-Host "    - Phase 2: enable services and finalize toolchains"
Write-Host ""

Write-Host "[*] Running Phase 1 inside WSL as root..."
Invoke-ArchPhase1 -DistroName $DistroName -DefaultUser $DefaultUser -DnsMode $DnsMode -RepoRoot $repoRoot

Write-Host "[*] Shutting down WSL to activate systemd and default user..."
& wsl.exe --shutdown

Write-Host "[*] Running Phase 2 inside WSL..."
Invoke-ArchPhase2 -DistroName $DistroName -DefaultUser $DefaultUser -DnsMode $DnsMode -RepoRoot $repoRoot

Write-Host "[*] Terminating distro before export..."
& wsl.exe --terminate $DistroName

Write-Host "[*] Exporting clean snapshot (warnings about sockets are harmless)..."
Export-ArchSnapshot -DistroName $DistroName -Output $snapshotFile
Write-Host "[+] Snapshot saved to: $snapshotFile"

Write-Host ""
Write-Host "Reset steps:"
Write-Host "  wsl --unregister $DistroName"
Write-Host "  wsl --import $DistroName C:\WSL\Arch $snapshotFile --version 2"
Write-Host ""
Write-Host "Done."