<#
    geoget.ps1 - Windows PowerShell deployment script for PC/GEOS Ensemble

    This script mirrors geoget.sh and prepares a runnable PC/GEOS Ensemble
    environment using the Basebox DOSBox-Staging fork on Windows. It downloads
    the latest builds, installs them under a chosen install root, copies a
    template Basebox configuration that mounts the Ensemble files, and creates
    ensemble launchers.

    standing in the geoget folder, use it from a classic cmd.exe window like this:
    .\geoget.cmd geosbbx2

    From a PowerShell prompt you can run:
    .\geoget.ps1 geosbbx2
#>

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# -----------------------------------------------------------------------------
# Configuration
# -----------------------------------------------------------------------------

$GEOS_RELEASE_URL = "https://github.com/bluewaysw/pcgeos/releases/download/CI-latest-issue-829/pcgeos-ensemble_nc.zip"
$BASEBOX_RELEASE_URL = "https://github.com/bluewaysw/pcgeos-basebox/releases/download/CI-latest-issue-13/pcgeos-basebox.zip"

$ScriptDir = Split-Path -Parent -Path $MyInvocation.MyCommand.Definition
$LocalUserConfigSource = Join-Path $ScriptDir 'templ/basebox.conf'
$LocalLauncherCmdTemplate = Join-Path $ScriptDir 'templ/ensemble.cmd'
$LocalLauncherShTemplate = Join-Path $ScriptDir 'templ/ensemble.sh'

$GeosArchiveRootName = 'ensemble'

function Write-Log {
    param(
        [Parameter(Mandatory = $true)][string]$Message
    )

    Write-Host "[geoget] $Message"
}

function Fail {
    param(
        [Parameter(Mandatory = $true)][string]$Message
    )

    Write-Error $Message
    exit 1
}

function Require-Command {
    param(
        [Parameter(Mandatory = $true)][string]$Name
    )

    if (-not (Get-Command -Name $Name -ErrorAction SilentlyContinue)) {
        Fail "Required command '$Name' not found. Please install it and re-run the script."
    }
}

function Get-UserHome {
    $userRoot = $env:USERPROFILE
    if ($userRoot) {
        return $userRoot
    }

    $homeEnv = $env:HOME
    if ($homeEnv) {
        return $homeEnv
    }

    Fail 'Neither USERPROFILE nor HOME environment variables are set.'
}

function Resolve-InstallRoot {
    param(
        [Parameter(Mandatory = $true)][string]$Root
    )

    $fullPath = if ([System.IO.Path]::IsPathRooted($Root)) {
        [System.IO.Path]::GetFullPath($Root)
    }
    else {
        $userRoot = Get-UserHome
        [System.IO.Path]::GetFullPath((Join-Path $userRoot $Root))
    }

    return $fullPath
}

function Download-File {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [Parameter(Mandatory = $true)][string]$Destination
    )

    Write-Log "Downloading $Url"
    Invoke-WebRequest -Uri $Url -OutFile $Destination
}

function Resolve-GeosArchiveRoot {
    param(
        [Parameter(Mandatory = $true)][string]$BaseDir
    )

    $defaultRoot = Join-Path $BaseDir $GeosArchiveRootName
    if (Test-Path -Path $defaultRoot -PathType Container) {
        return $defaultRoot
    }

    $geosIni = Get-ChildItem -Path $BaseDir -Filter 'geos.ini' -File -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($geosIni) {
        return $geosIni.Directory.FullName
    }

    $ensembleDir = Get-ChildItem -Path $BaseDir -Filter 'ensemble' -Directory -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($ensembleDir) {
        return $ensembleDir.FullName
    }

    return ''
}

function Select-BaseboxBinary {
    param(
        [Parameter(Mandatory = $true)][string]$BaseboxRoot
    )

    $platform = [System.Runtime.InteropServices.RuntimeInformation]
    $preferred = @()

    switch ($true) {
        { $platform::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows) } {
            $preferred += 'binnt/basebox.exe'
            break
        }
        { $platform::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Linux) } {
            $preferred += @('binl64/basebox', 'binl/basebox')
            break
        }
        { $platform::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::OSX) } {
            $preferred += 'binmac/basebox'
            break
        }
    }

    $candidates = $preferred + @('binl64/basebox', 'binl/basebox', 'binmac/basebox', 'binnt/basebox.exe')

    foreach ($relative in $candidates | Select-Object -Unique) {
        $candidate = Join-Path $BaseboxRoot $relative
        if (Test-Path -Path $candidate -PathType Leaf) {
            return $candidate
        }
    }

    return ''
}

if ($args.Count -lt 1) {
    Fail "Usage: pwsh -File geoget.ps1 <install-root>"
}

$InstallRoot = Resolve-InstallRoot -Root $args[0]
$DriveCDir = Join-Path $InstallRoot 'drivec'
$GeosInstallDir = Join-Path $DriveCDir 'ensemble'
$BaseboxDir = Join-Path $InstallRoot 'basebox'
$LocalLauncherCmd = Join-Path $InstallRoot 'ensemble.cmd'
$LocalLauncherSh = Join-Path $InstallRoot 'ensemble.sh'

function Prepare-Environment {
    Write-Log 'Checking prerequisites'
    Require-Command -Name 'Invoke-WebRequest'
    Require-Command -Name 'Expand-Archive'

    if (Test-Path -Path $InstallRoot) {
        Write-Log "Removing existing installation at $InstallRoot"
        Remove-Item -Path $InstallRoot -Recurse -Force
    }

    Write-Log "Preparing installation directories under $InstallRoot"
    New-Item -ItemType Directory -Path $GeosInstallDir -Force | Out-Null
    New-Item -ItemType Directory -Path $BaseboxDir -Force | Out-Null
}

function Extract-Archives {
    $tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        $geosZip = Join-Path $tempDir 'pcgeos-ensemble.zip'
        $baseboxZip = Join-Path $tempDir 'pcgeos-basebox.zip'

        Download-File -Url $GEOS_RELEASE_URL -Destination $geosZip
        Download-File -Url $BASEBOX_RELEASE_URL -Destination $baseboxZip

        Write-Log 'Extracting Ensemble archive'
        Expand-Archive -Path $geosZip -DestinationPath (Join-Path $tempDir 'ensemble') -Force

        Write-Log 'Extracting Basebox archive'
        Expand-Archive -Path $baseboxZip -DestinationPath (Join-Path $tempDir 'basebox') -Force

        Write-Log "Installing Ensemble into $GeosInstallDir"
        $geosSource = Resolve-GeosArchiveRoot -BaseDir (Join-Path $tempDir 'ensemble')
        if (-not $geosSource -or -not (Test-Path -Path $geosSource -PathType Container)) {
            Fail "Unable to locate Ensemble archive root inside $tempDir/ensemble."
        }

        Copy-Item -Path (Join-Path $geosSource '*') -Destination $GeosInstallDir -Recurse -Force

        Write-Log "Installing Basebox into $BaseboxDir"
        $baseboxSource = Join-Path $tempDir 'basebox/pcgeos-basebox/*'
        Copy-Item -Path $baseboxSource -Destination $BaseboxDir -Recurse -Force

        $runtimeInfo = [System.Runtime.InteropServices.RuntimeInformation]
        if (-not $runtimeInfo::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows)) {
            if (Get-Command -Name 'chmod' -ErrorAction SilentlyContinue) {
                Write-Log 'Ensuring Basebox executables are marked executable'
                Get-ChildItem -Path $BaseboxDir -Recurse -File -Include 'basebox', 'basebox.exe', '*.sh' |
                    ForEach-Object {
                        try {
                            & chmod +x -- $_.FullName
                        }
                        catch {
                            Write-Log "Warning: Failed to mark $($_.FullName) as executable: $($_.Exception.Message)"
                        }
                    }
            }
            else {
                Write-Log 'Warning: chmod not available; skipping executable bit adjustments.'
            }
        }

        $detected = Select-BaseboxBinary -BaseboxRoot $BaseboxDir
        if (-not $detected) {
            Fail "Unable to locate the Basebox executable inside $BaseboxDir"
        }
    }
    finally {
        if (Test-Path -Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force
        }
    }
}

function Copy-LocalUserConfig {
    $destination = Join-Path $BaseboxDir 'basebox.conf'
    if (-not (Test-Path -Path $LocalUserConfigSource -PathType Leaf)) {
        Write-Log "Warning: Local Basebox config template missing at $LocalUserConfigSource"
        return
    }

    Write-Log "Copying local Basebox user config from $LocalUserConfigSource"
    Copy-Item -Path $LocalUserConfigSource -Destination $destination -Force

    $drivecPath = [System.IO.Path]::GetFullPath($DriveCDir)
    (Get-Content -Path $destination -Raw) -replace '\{\{TAG\}\}', $drivecPath |
        Set-Content -Path $destination -Encoding ASCII
}

function Create-Launcher {
    Write-Log "Creating Ensemble launchers at $LocalLauncherCmd and $LocalLauncherSh"

    if (Test-Path -Path $LocalLauncherCmdTemplate -PathType Leaf) {
        Copy-Item -Path $LocalLauncherCmdTemplate -Destination $LocalLauncherCmd -Force
    }
    else {
        Write-Log "Warning: Launcher template missing at $LocalLauncherCmdTemplate"
    }

    if (Test-Path -Path $LocalLauncherShTemplate -PathType Leaf) {
        Copy-Item -Path $LocalLauncherShTemplate -Destination $LocalLauncherSh -Force
    }
    else {
        Write-Log "Warning: Launcher template missing at $LocalLauncherShTemplate"
    }

    $runtimeInfo = [System.Runtime.InteropServices.RuntimeInformation]
    if (-not $runtimeInfo::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows)) {
        if (Get-Command -Name 'chmod' -ErrorAction SilentlyContinue) {
            try {
                & chmod +x -- $LocalLauncherSh
            }
            catch {
                Write-Log "Warning: Failed to mark $LocalLauncherSh as executable: $($_.Exception.Message)"
            }
        }
        else {
            Write-Log 'Warning: chmod not available; launcher may not be executable.'
        }
    }
}

function Main {
    Prepare-Environment
    Extract-Archives
    Copy-LocalUserConfig
    Create-Launcher

    Write-Log 'Deployment complete.'
}

Main
