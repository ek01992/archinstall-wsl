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
# TTY-aware theming / TUI helpers
# -------------------------------
$IsTty = $true
try { $IsTty = -not [Console]::IsOutputRedirected } catch { $IsTty = $false }
if ($IsTty -and $env:NO_COLOR) { $PSStyle.OutputRendering = 'PlainText' }

# Muted color detection
$fgProps = @()
try { $fgProps = $PSStyle.Foreground.PSObject.Properties.Name } catch { $fgProps = @() }
$MutedColor =
  if ($fgProps -contains 'BrightBlack') { $PSStyle.Foreground.BrightBlack }
  elseif ($fgProps -contains 'DarkGray') { $PSStyle.Foreground.DarkGray }
  else { '' }

$Theme = @{
  Accent = $PSStyle.Foreground.Cyan
  Good   = $PSStyle.Foreground.Green
  Warn   = $PSStyle.Foreground.Yellow
  Bad    = $PSStyle.Foreground.Red
  Muted  = $MutedColor
  Bold   = $PSStyle.Bold
  Reset  = $PSStyle.Reset
}

function Get-TuiWidth { try { [Math]::Max(20, [Console]::WindowWidth) } catch { 80 } }

function Get-TuiChars {
  $isUtf = $true
  try { $isUtf = [Console]::OutputEncoding.BodyName -match 'utf' } catch { $isUtf = $true }
  if ($env:UI_ASCII -eq '1' -or -not $isUtf) {
    return @{ TL='+'; TR='+'; BL='+'; BR='+'; H='-'; V='|' }
  } else {
    return @{ TL='┌'; TR='┐'; BL='└'; BR='┘'; H='─'; V='│' }
  }
}

function Write-Section {
  param([Parameter(Mandatory)][string]$Title)
  $w = Get-TuiWidth
  $ch = Get-TuiChars
  $hr = $ch.H * ([Math]::Max(10, $w - 2))
  Write-Host "$($Theme.Muted)$($ch.TL)$hr$($ch.TR)$($Theme.Reset)"
  Write-Host "$($Theme.Bold)$($Theme.Accent)$($ch.V) $Title$($Theme.Reset)"
  Write-Host "$($Theme.Muted)$($ch.BL)$hr$($ch.BR)$($Theme.Reset)"
}

function Write-StatusEx {
  param(
    [Parameter(Mandatory)][ValidateSet('Info','Ok','Warn','Error')]$Level,
    [Parameter(Mandatory)][string]$Text
  )
  switch ($Level) {
    'Info'  { Write-Host "$($Theme.Accent)[*]$($Theme.Reset) $Text" }
    'Ok'    { Write-Host "$($Theme.Good)[+]$($Theme.Reset) $Text" }
    'Warn'  { Write-Host "$($Theme.Warn)[!]$($Theme.Reset) $Text" }
    'Error' { Write-Host "$($Theme.Bad)[x]$($Theme.Reset) $Text" }
  }
}

function Invoke-Step {
  [CmdletBinding()]
  param([Parameter(Mandatory)][string]$Activity, [Parameter(Mandatory)][scriptblock]$Script)
  $id = (Get-Random -Minimum 1000 -Maximum 9999)
  $sw = [Diagnostics.Stopwatch]::StartNew()
  Write-Progress -Id $id -Activity $Activity -Status 'Working...' -PercentComplete 0
  try {
    & $Script
    $sw.Stop()
    Write-Progress -Id $id -Completed -Activity $Activity
    Write-StatusEx -Level Ok -Text "$Activity ($([int]$sw.Elapsed.TotalSeconds)s)"
  } catch {
    $sw.Stop()
    Write-Progress -Id $id -Completed -Activity $Activity
    Write-StatusEx -Level Error -Text "$Activity failed: $($_.Exception.Message)"
    throw
  }
}

function Show-TuiTable {
  param([Parameter(Mandatory, ValueFromPipeline)]$InputObject)
  process {
    $pairs =
      if ($InputObject -is [hashtable]) { $InputObject.GetEnumerator() | Sort-Object Name }
      else { $InputObject.PSObject.Properties | Sort-Object Name }
    $max = ($pairs | ForEach-Object { $_.Name.Length } | Measure-Object -Maximum).Maximum
    foreach ($p in $pairs) {
      $k = $p.Name; $v = $p.Value
      $fmt = "{0}{1,-" + $max + "}{2} : {3}"
      Write-Host ($fmt -f $Theme.Muted, $k, $Theme.Reset, $v)
    }
  }
}

function Read-TuiChoice {
  param(
    [Parameter(Mandatory)][string]$Title,
    [Parameter(Mandatory)][string]$Message,
    [Parameter(Mandatory)][string[]]$Choices,
    [int]$DefaultIndex = 0
  )
  $cd = foreach ($c in $Choices) { New-Object Management.Automation.Host.ChoiceDescription "&$c", $c }
  $sel = $host.UI.PromptForChoice($Title, $Message, $cd, $DefaultIndex)
  return $Choices[$sel]
}

function Read-TuiConfirm {
  param([Parameter(Mandatory)][string]$Message, [switch]$DefaultYes)
  $choices = @('Yes','No'); $def = if ($DefaultYes) { 0 } else { 1 }
  (Read-TuiChoice -Title 'Confirm' -Message $Message -Choices $choices -DefaultIndex $def) -eq 'Yes'
}

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
    Write-StatusEx -Level Info -Text "WSL default version set to 2."
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
    [byte[]]$existing = [System.IO.File]::ReadAllBytes($Path)
    $newBytes = [System.Text.Encoding]::GetEncoding($Encoding).GetBytes($NewContent)
    $areEqual = [System.Linq.Enumerable]::SequenceEqual($existing, $newBytes)
    $shouldWrite = -not $areEqual
    if ($shouldWrite -and $PSCmdlet.ShouldProcess($Path, "Backup to ${Path}${BackupSuffix}")) {
      Copy-Item $Path "${Path}${BackupSuffix}" -Force
      Write-StatusEx -Level Info -Text "Backed up existing file to ${Path}${BackupSuffix}"
    }
  }
  if ($shouldWrite -and $PSCmdlet.ShouldProcess($Path, "Write updated content")) {
    $NewContent | Set-Content -Encoding $Encoding -NoNewline -Path $Path
    Write-StatusEx -Level Ok -Text "Wrote $Path"
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

function ConvertTo-Lf {
  param([Parameter(Mandatory)][string]$Path)
  $text = Get-Content -Raw -Encoding UTF8 $Path
  $text = $text -replace "`r`n","`n"
  $text = $text -replace "`r","`n"
  $temp = Join-Path $env:TEMP ([IO.Path]::GetFileNameWithoutExtension($Path) + "-lf" + [IO.Path]::GetExtension($Path))
  Set-Content -Path $temp -Value $text -NoNewline -Encoding utf8NoBOM
  Write-StatusEx -Level Ok -Text "Prepared $temp (LF line endings)."
  return $temp
}
Set-Alias -Name Convert-ToLF -Value ConvertTo-Lf -Scope Local -ErrorAction SilentlyContinue

function Prepare-BootstrapScript {
  return (ConvertTo-Lf -Path $bootstrapLocal)
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
    Write-StatusEx -Level Warn -Text "Unregistering '$Name'..."
    & wsl.exe --terminate $Name | Out-Null
    & wsl.exe --unregister $Name
  }
}

function Install-WslDistributionNoLaunch {
  param([Parameter(Mandatory)][string]$Name)
  if ($PSCmdlet.ShouldProcess("WSL distro '$Name'", "Install --no-launch")) {
    Write-StatusEx -Level Info -Text "Installing '$Name'..."
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

function Initialize-WslDistribution {
  param(
    [Parameter(Mandatory)][string]$Name,
    [switch]$ForceReinstall
  )
  Ensure-WslDistributionPresent -Name $Name -ForceReinstall:$ForceReinstall
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
    Write-StatusEx -Level Info -Text "Exporting clean snapshot (warnings about sockets are harmless)..."
    & wsl.exe --export $DistroName $snapshotFile
    Write-StatusEx -Level Ok -Text "Snapshot saved to: $snapshotFile"
  }
}

function Show-Preflight {
  $obj = [pscustomobject]@{
    Distro         = $DistroName
    'Default user' = $DefaultUser
    CPUs           = $WSLCPUs
    Memory         = $WSLMemory
    SwapGB         = $SwapGB
    'DNS mode'     = $DnsMode
    'Snapshot file'= $snapshotFile
  }
  $obj | Show-TuiTable
}

function Show-Plan {
  Write-StatusEx -Level Info -Text "Phase 1: base system, keyring, packages, user, DNS($DnsMode), dotfiles"
  Write-StatusEx -Level Info -Text "Restart WSL to enable systemd & default user"
  Write-StatusEx -Level Info -Text "Phase 2: enable services and finalize toolchains"
  Write-Host ""
}

function Main {
  Write-Section "Preflight summary"
  Show-Preflight
  Ensure-ExecutionPolicy

  if (-not $PSBoundParameters.ContainsKey('DnsMode') -and $IsTty) {
    $DnsMode = Read-TuiChoice -Title "DNS mode" -Message "Select DNS strategy" -Choices @('static','resolved','wsl') -DefaultIndex 0
  }

  if (-not (Test-WslAvailable)) {
    throw "WSL is not installed or not available. Install WSL and WSL2 first."
  }
  Write-StatusEx -Level Ok -Text "WSL is installed and available."

  Write-Section "Applying host configuration"
  Invoke-Step -Activity "Set WSL default version 2" -Script { Set-WslDefaultVersion2 }
  Invoke-Step -Activity "Write .wslconfig" -Script { Set-WslConfigFile }
  Write-StatusEx -Level Info -Text "Applying .wslconfig (wsl --shutdown)..."
  Restart-Wsl
  Start-Sleep -Milliseconds 1500

  Write-StatusEx -Level Info -Text "Checking WSL installation..."
  & wsl.exe --status | Out-Null

  Write-Section "Plan"
  Show-Plan

  $tempBootstrap = Prepare-BootstrapScript
  if ($tempBootstrap.StartsWith("\\") -or $repoRoot.StartsWith("\\")) {
    throw "Repository or temp path is a UNC path. Please use a local drive path for correct /mnt mapping."
  }
  $mntBootstrap = ConvertTo-WslPath -WindowsPath $tempBootstrap
  $mntRepoRoot  = ConvertTo-WslPath -WindowsPath $repoRoot

  Write-Section "Phase 1"
  Invoke-Step -Activity "Prepare distro '$DistroName'" -Script { Ensure-WslDistributionPresent -Name $DistroName -ForceReinstall:$Force }
  Invoke-Step -Activity "Run Phase 1 inside WSL" -Script { Invoke-WslPhase -Phase 'phase1' -MntBootstrap $mntBootstrap -MntRepoRoot $mntRepoRoot }

  Write-Section "Phase 2"
  Write-StatusEx -Level Info -Text "Shutting down WSL to activate systemd and default user..."
  Restart-Wsl
  Invoke-Step -Activity "Run Phase 2 inside WSL" -Script { Invoke-WslPhase -Phase 'phase2' -MntBootstrap $mntBootstrap -MntRepoRoot $mntRepoRoot }

  Write-Section "Snapshot"
  Invoke-Step -Activity "Export clean snapshot" -Script { Export-WslSnapshot }

  Write-Section "Summary"
  Write-StatusEx -Level Info -Text "Reset steps:"
  Write-Host "  wsl --unregister $DistroName"
  Write-Host "  wsl --import $DistroName C:\WSL\Arch $snapshotFile --version 2"
  Write-Host ""
  Write-StatusEx -Level Ok -Text "Done."
}

Main