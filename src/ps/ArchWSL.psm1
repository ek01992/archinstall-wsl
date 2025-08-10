# ArchWSL PowerShell module: WSL lifecycle and bootstrap orchestration
# Exports: New-ArchDistro, Invoke-ArchPhase1, Invoke-ArchPhase2, Export-ArchSnapshot, Set-ArchWslConfig, Get-ArchDistroList

using namespace System.IO

function Get-ArchDistroList {
  & wsl.exe --list --quiet | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne '' }
}

function Set-ArchWslConfig {
  param(
    [string]$Memory = $env:WSL_MEMORY ? $env:WSL_MEMORY : '8GB',
    [string]$Processors = $env:WSL_CPUS ? $env:WSL_CPUS : '4',
    [int]$SwapGB = 4
  )
  $path = Join-Path $env:USERPROFILE ".wslconfig"
  $content = @"
[wsl2]
memory=$Memory
processors=$Processors
swap=$(${SwapGB})GB
localhostForwarding=true
"@
  if (Test-Path $path) {
    $existing = Get-Content -Raw $path
    if ($existing -ne $content) { Copy-Item $path "$path.bak" -Force }
  }
  $content | Set-Content -NoNewline -Encoding ASCII $path
  Write-Host "[*] Wrote $path (memory=$Memory, processors=$Processors, swap=${SwapGB}GB)"
}

function New-ArchDistro {
  param(
    [string]$DistroName = 'archlinux',
    [switch]$Force
  )
  $null = & wsl.exe --status 2>$null
  if ($LASTEXITCODE -ne 0) { throw "WSL is not installed." }
  try { $null = & wsl.exe --set-default-version 2 } catch {}

  $installed = Get-ArchDistroList
  if ($installed -contains $DistroName) {
    if (-not $Force) { throw "Distro '$DistroName' exists. Use -Force to replace." }
    & wsl.exe --terminate $DistroName | Out-Null
    & wsl.exe --unregister $DistroName
  }

  & wsl.exe --install --no-launch -d $DistroName
}

function Invoke-ArchPhase1 {
  param(
    [string]$DistroName = 'archlinux',
    [string]$DefaultUser = 'erik',
    [ValidateSet('static','resolved','wsl')]
    [string]$DnsMode = 'static',
    [string]$RepoRoot = (Split-Path -Parent $PSScriptRoot)
  )
  $bin = Join-Path $RepoRoot "bin/arch-wsl"
  if (-not (Test-Path $bin)) { throw "bin/arch-wsl not found at $bin" }

  $tmp = Join-Path $env:TEMP "arch-wsl-bootstrap.sh"
  Copy-Item $bin $tmp -Force

  if ($tmp.StartsWith("\\") -or $RepoRoot.StartsWith("\\")) {
    throw "Repository or temp path is a UNC path. Use a local drive path."
  }

  $mntPath = "/mnt/" + ($tmp.Substring(0,1).ToLower()) + ($tmp.Substring(2) -replace "\\","/")
  $mntRepoRoot = "/mnt/" + ($RepoRoot.Substring(0,1).ToLower()) + ($RepoRoot.Substring(2) -replace "\\","/")

  & wsl.exe -d $DistroName -u root -- bash -lc "chmod +x '$mntPath' && DEFAULT_USER='$DefaultUser' REPO_ROOT_MNT='$mntRepoRoot' DNS_MODE='$DnsMode' '$mntPath' phase1"
}

function Invoke-ArchPhase2 {
  param(
    [string]$DistroName = 'archlinux',
    [string]$DefaultUser = 'erik',
    [ValidateSet('static','resolved','wsl')]
    [string]$DnsMode = 'static',
    [string]$RepoRoot = (Split-Path -Parent $PSScriptRoot)
  )
  $bin = Join-Path $RepoRoot "bin/arch-wsl"
  if (-not (Test-Path $bin)) { throw "bin/arch-wsl not found at $bin" }
  $tmp = Join-Path $env:TEMP "arch-wsl-bootstrap.sh"
  Copy-Item $bin $tmp -Force
  $mntPath = "/mnt/" + ($tmp.Substring(0,1).ToLower()) + ($tmp.Substring(2) -replace "\\","/")
  & wsl.exe -d $DistroName -u root -- bash -lc "DEFAULT_USER='$DefaultUser' DNS_MODE='$DnsMode' '$mntPath' phase2"
}

function Export-ArchSnapshot {
  param(
    [string]$DistroName = 'archlinux',
    [string]$Output = (Join-Path $env:USERPROFILE "arch.tar.bak")
  )
  & wsl.exe --terminate $DistroName
  & wsl.exe --export $DistroName $Output
  Write-Host "[+] Snapshot saved to: $Output"
}

Export-ModuleMember -Function *-Arch*
