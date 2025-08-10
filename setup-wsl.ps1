# setup-wsl.ps1
# Requires Windows 10 2004+ or Windows 11 with WSL2 enabled

[CmdletBinding(SupportsShouldProcess = $true, ConfirmImpact = 'Medium')]
param(
    [switch]$Force,

    [ValidatePattern('^[A-Za-z0-9._-]+$')]
    [string]$DistroName = 'archlinux',

    [ValidatePattern('^[A-Za-z_][A-Za-z0-9_-]*$')]
    [string]$DefaultUser = 'erik',

    [ValidateSet('static','resolved','wsl')]
    [string]$DnsMode = 'static',

    [ValidateRange(0, 256)]
    [int]$SwapGB = 4
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# -------------------------------
# Configuration
# -------------------------------
$repoRoot     = Split-Path -Parent $MyInvocation.MyCommand.Path
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
  param([Parameter(Mandatory)][string]$Name)
  $d = Get-WslDistributions
  return ($d -contains $Name)
}

function Wait-WslDistribution {
  param(
    [Parameter(Mandatory)][string]$Name,
    [int]$TimeoutSec = 60
  )
  $sw = [Diagnostics.Stopwatch]::StartNew()
  while ($sw.Elapsed.TotalSeconds -lt $TimeoutSec) {
    if (Test-WslDistributionExists -Name $Name) { return $true }
    Start-Sleep -Seconds 2
  }
  return $false
}

function Get-WslConfigContent {
  @"
[wsl2]
memory=$WSLMemory
processors=$WSLCPUs
swap=$(${SwapGB})GB
localhostForwarding=true
"@
}

function Set-FileContentIfChanged {
  param(
    [Parameter(Mandatory)][string]$Path,
    [Parameter(Mandatory)][string]$NewContent,
    [string]$BackupSuffix = '.bak',
    [string]$Encoding = 'ASCII'
  )
  $shouldWrite = $true
  if (Test-Path $Path) {
    # Use .NET to read raw bytes reliably across PS versions.
    [byte[]]$existing = [System.IO.File]::ReadAllBytes($Path)
    $newBytes = [System.Text.Encoding]::GetEncoding($Encoding).GetBytes($NewContent)
    $areEqual = [System.Linq.Enumerable]::SequenceEqual($existing, $newBytes)
    $shouldWrite = -not $areEqual
    if ($shouldWrite -and $PSCmdlet.ShouldProcess($Path, "Backup to ${Path}${BackupSuffix}")) {
      Copy-Item $Path "${Path}${BackupSuffix}" -Force
      Write-Status "[i] Backed up existing file to ${Path}${BackupSuffix}" -Color DarkGray
    }
  }
  if ($shouldWrite -and $PSCmdlet.ShouldProcess($Path, "Write updated content")) {
    $NewContent | Set-Content -Encoding $Encoding -NoNewline -Path $Path
    Write-Status "[*] Wrote $Path" -Color Cyan
  }
}

function Set-WslConfigFile {
  $content = Get-WslConfigContent
  Set-FileContentIfChanged -Path $wslConfigPath -NewContent $content -Encoding 'ASCII'
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

function Convert-ToLF {
  param([Parameter(Mandatory)][string]$Path)
  $text = Get-Content -Raw -Encoding UTF8 $Path
  $text = $text -replace "`r`n","`n"
  $text = $text -replace "`r","`n"
  $temp = Join-Path $env:TEMP ([IO.Path]::GetFileNameWithoutExtension($Path) + "-lf" + [IO.Path]::GetExtension($Path))
  Set-Content -Path $temp -Value $text -NoNewline -Encoding utf8NoBOM
  Write-Status "[*] Prepared $temp (LF line endings)." -Color Cyan
  return $temp
}

function Prepare-BootstrapScript {
  return (Convert-ToLF -Path $bootstrapLocal)
}

function Invoke-WslChecked {
  param(
    [Parameter(Mandatory)][string[]]$ArgList,
    [Parameter(Mandatory)][string]$ErrorContext
  )
  & wsl.exe @ArgList
  if ($LASTEXITCODE -ne 0) {
    throw "$ErrorContext failed with exit code $LASTEXITCODE"
  }
}

function Remove-WslDistribution {
  param([Parameter(Mandatory)][string]$Name)
  if ($PSCmdlet.ShouldProcess("WSL distro '$Name'", "Unregister")) {
    Write-Status "[*] Unregistering '$Name'..." -Color Yellow
    & wsl.exe --terminate $Name | Out-Null
    & wsl.exe --unregister $Name
  }
}

function Install-WslDistributionNoLaunch {
  param([Parameter(Mandatory)][string]$Name)
  if ($PSCmdlet.ShouldProcess("WSL distro '$Name'", "Install --no-launch")) {
    Write-Status "[*] Installing '$Name'..." -Color Cyan
    try {
      & wsl.exe --install --no-launch -d $Name
    } catch {
      Write-Warning "wsl.exe --install failed. If the Microsoft Store Arch is unavailable, consider importing an Arch rootfs manually."
      throw
    }
    if (-not (Wait-WslDistribution -Name $Name -TimeoutSec 120)) {
      throw "Distro '$Name' did not appear after install."
    }
  }
}

function Ensure-WslDistributionPresent {
  param(
    [Parameter(Mandatory)][string]$Name,
    [switch]$ForceReinstall
  )
  if (Test-WslDistributionExists -Name $Name) {
    if (-not $ForceReinstall) {
      $resp = Read-Host "Distro '$Name' exists. Unregister it and reinstall? (y/N)"
      if ($resp -notin @("y","Y")) {
        return
      }
    }
    Remove-WslDistribution -Name $Name
  }
  Install-WslDistributionNoLaunch -Name $Name
}

function New-PhaseCommand {
  param(
    [Parameter(Mandatory)][ValidateSet('phase1','phase2')] [string]$Phase,
    [Parameter(Mandatory)][string]$MntBootstrap,
    [Parameter(Mandatory)][string]$MntRepoRoot
  )
  if ($Phase -eq 'phase1') {
    return "chmod +x '$MntBootstrap' && DEFAULT_USER='$DefaultUser' WSL_MEMORY='$WSLMemory' WSL_CPUS='$WSLCPUs' REPO_ROOT_MNT='$MntRepoRoot' DNS_MODE='$DnsMode' '$MntBootstrap' phase1"
  } else {
    return "DEFAULT_USER='$DefaultUser' DNS_MODE='$DnsMode' '$MntBootstrap' phase2"
  }
}

function Invoke-WslPhase {
  param(
    [Parameter(Mandatory)][ValidateSet('phase1','phase2')] [string]$Phase,
    [Parameter(Mandatory)][string]$MntBootstrap,
    [Parameter(Mandatory)][string]$MntRepoRoot
  )
  $cmd = New-PhaseCommand -Phase $Phase -MntBootstrap $MntBootstrap -MntRepoRoot $MntRepoRoot
  $context = if ($Phase -eq 'phase1') { "Run Phase 1 as root" } else { "Run Phase 2" }
  if ($PSCmdlet.ShouldProcess("WSL '$DistroName'", $context)) {
    Invoke-WslChecked -ArgList @('-d', $DistroName, '-u', 'root', '--', 'bash', '-lc', $cmd) -ErrorContext "Phase $Phase"
  }
}

function Restart-Wsl {
  if ($PSCmdlet.ShouldProcess("WSL", "Shutdown")) {
    & wsl.exe --shutdown
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
  Restart-Wsl
  Start-Sleep -Milliseconds 1500

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
  Ensure-WslDistributionPresent -Name $DistroName -ForceReinstall:$Force

  Write-Status "[*] Running Phase 1 inside WSL as root..." -Color Cyan
  Invoke-WslPhase -Phase 'phase1' -MntBootstrap $mntBootstrap -MntRepoRoot $mntRepoRoot

  Write-Status "[*] Shutting down WSL to activate systemd and default user..." -Color Cyan
  Restart-Wsl

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