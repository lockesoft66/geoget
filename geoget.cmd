#!/usr/bin/env pwsh
<#!
    geoget.cmd - Windows PowerShell deployment script for PC/GEOS Ensemble

    This script mirrors geoget.sh and prepares a runnable PC/GEOS Ensemble
    environment using the Basebox DOSBox-Staging fork on Windows. It downloads
    the latest builds, installs them under a chosen install root, generates a
    Basebox configuration that mounts the Ensemble files, and creates an
    ensemble.cmd launcher.
!>

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# -----------------------------------------------------------------------------
# Configuration
# -----------------------------------------------------------------------------

$GEOS_RELEASE_URL = "https://github.com/bluewaysw/pcgeos/releases/download/CI-latest-issue-829/pcgeos-ensemble_nc.zip"
$BASEBOX_RELEASE_URL = "https://github.com/bluewaysw/pcgeos-basebox/releases/download/CI-latest-issue-13/pcgeos-basebox.zip"

$ScriptDir = Split-Path -Parent -Path $MyInvocation.MyCommand.Definition
$LocalUserConfigSource = Join-Path $ScriptDir 'basebox.conf'

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

function Resolve-InstallRoot {
    param(
        [Parameter(Mandatory = $true)][string]$Root
    )

    if ([System.IO.Path]::IsPathRooted($Root)) {
        return (Get-Item -Path $Root).FullName
    }

    $userRoot = $env:USERPROFILE
    if (-not $userRoot) {
        Fail 'USERPROFILE environment variable is not set.'
    }

    return (Join-Path $userRoot $Root)
}

function Download-File {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [Parameter(Mandatory = $true)][string]$Destination
    )

    Write-Log "Downloading $Url"
    Invoke-WebRequest -Uri $Url -OutFile $Destination -UseBasicParsing
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

if ($args.Count -lt 1) {
    Fail "Usage: pwsh -File geoget.cmd <install-root>"
}

$InstallRoot = Resolve-InstallRoot -Root $args[0]
$DriveCDir = Join-Path $InstallRoot 'drivec'
$GeosInstallDir = Join-Path $DriveCDir 'ensemble'
$BaseboxDir = Join-Path $InstallRoot 'basebox'
$BaseboxBaseConfig = Join-Path $BaseboxDir 'basebox-geos.conf'
$LocalLauncher = Join-Path $InstallRoot 'ensemble.cmd'

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
        Copy-Item -Path (Join-Path (Join-Path $tempDir 'basebox') 'pcgeos-basebox' '*') -Destination $BaseboxDir -Recurse -Force
    }
    finally {
        if (Test-Path -Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force
        }
    }
}

function Create-BaseboxConfig {
    $baseboxExe = Join-Path $BaseboxDir 'binnt/basebox.exe'
    if (-not (Test-Path -Path $baseboxExe -PathType Leaf)) {
        Fail "Unable to locate Basebox executable at $baseboxExe"
    }

    Write-Log 'Generating Basebox configuration'

    $xdgRoot = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
    New-Item -ItemType Directory -Path $xdgRoot -Force | Out-Null
    $xdgConfig = Join-Path $xdgRoot 'config'
    New-Item -ItemType Directory -Path $xdgConfig -Force | Out-Null

    $previousXdg = $env:XDG_CONFIG_HOME
    $env:XDG_CONFIG_HOME = $xdgConfig
    try {
        $printConf = & $baseboxExe --printconf 2>$null
        $configLine = ($printConf | Where-Object { $_ -match '\S' } | Select-Object -Last 1)
        if (-not $configLine) {
            Fail 'Failed to determine the Basebox configuration path via --printconf.'
        }

        $configLine = $configLine.TrimEnd("`r")
        if ($configLine -match ':') {
            $configPath = ($configLine -split ':', 2)[1].Trim()
        }
        else {
            $configPath = $configLine.Trim()
        }

        if (-not $configPath) {
            Fail 'Unable to parse the Basebox configuration path from --printconf output.'
        }

        $configDir = Split-Path -Parent -Path $configPath
        New-Item -ItemType Directory -Path $configDir -Force | Out-Null
        if (Test-Path -Path $configPath) {
            Remove-Item -Path $configPath -Force
        }

        & $baseboxExe -c exit *> $null

        if (-not (Test-Path -Path $configPath -PathType Leaf)) {
            Fail "Basebox did not create a configuration file at $configPath."
        }

        $configLines = Get-Content -Path $configPath -ErrorAction Stop
        $drivecPath = [System.IO.Path]::GetFullPath($DriveCDir)
        $autoexecBlock = @(
            '@echo off',
            "mount c \"$drivecPath\" -t dir",
            'c:',
            'cd ensemble',
            'loader',
            'exit'
        )

        $outputLines = New-Object System.Collections.Generic.List[string]
        $inAutoexec = $false
        $autoexecInserted = $false

        foreach ($line in $configLines) {
            $trimmed = $line.Trim()
            if ($trimmed -match '^\[autoexec\]\s*$') {
                $outputLines.Add($line)
                foreach ($autoLine in $autoexecBlock) {
                    $outputLines.Add($autoLine)
                }
                $inAutoexec = $true
                $autoexecInserted = $true
                continue
            }

            if ($inAutoexec) {
                if ($trimmed -match '^\[') {
                    $outputLines.Add($line)
                    $inAutoexec = $false
                }
                continue
            }

            $outputLines.Add($line)
        }

        if (-not $autoexecInserted) {
            $outputLines.Add('[autoexec]')
            foreach ($autoLine in $autoexecBlock) {
                $outputLines.Add($autoLine)
            }
        }

        Set-Content -Path $BaseboxBaseConfig -Value $outputLines -Encoding UTF8 -NoNewline:$false
    }
    finally {
        $env:XDG_CONFIG_HOME = $previousXdg
        if (Test-Path -Path $xdgRoot) {
            Remove-Item -Path $xdgRoot -Recurse -Force
        }
    }
}

function Copy-LocalUserConfig {
    $destination = Join-Path $BaseboxDir 'basebox.conf'
    if (Test-Path -Path $LocalUserConfigSource -PathType Leaf) {
        Write-Log "Copying local Basebox user config from $LocalUserConfigSource"
        Copy-Item -Path $LocalUserConfigSource -Destination $destination -Force
    }
}

function Create-Launcher {
    Write-Log "Creating Ensemble launcher at $LocalLauncher"

    $launcherContent = @'
@echo off
setlocal
set SCRIPT_DIR=%~dp0
set BASEBOX_DIR=%SCRIPT_DIR%basebox
set BASEBOX_EXEC=%BASEBOX_DIR%\binnt\basebox.exe
set BASE_CONFIG_FILE=%BASEBOX_DIR%\basebox-geos.conf
set USER_CONFIG_FILE=%BASEBOX_DIR%\basebox.conf

if not exist "%BASEBOX_EXEC%" (
    echo Error: Unable to locate the Basebox executable at %BASEBOX_EXEC%.
    exit /b 1
)

if not exist "%BASE_CONFIG_FILE%" (
    echo Error: Missing Basebox configuration at %BASE_CONFIG_FILE%.
    exit /b 1
)

"%BASEBOX_EXEC%" -conf "%BASE_CONFIG_FILE%" -conf "%USER_CONFIG_FILE%" %*
'@

    Set-Content -Path $LocalLauncher -Value $launcherContent -Encoding ASCII -NoNewline:$false
}

function Main {
    Prepare-Environment
    Extract-Archives
    Copy-LocalUserConfig
    Create-BaseboxConfig
    Create-Launcher

    Write-Log 'Deployment complete.'
}

Main
