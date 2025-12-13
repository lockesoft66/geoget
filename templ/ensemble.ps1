$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$BaseboxDir = Join-Path $ScriptDir "basebox"
$UserConfig = Join-Path $BaseboxDir "basebox.conf"

$Candidates = @(
    Join-Path $BaseboxDir "binnt/basebox.exe",
    Join-Path $BaseboxDir "binl64/basebox",
    Join-Path $BaseboxDir "binl/basebox",
    Join-Path $BaseboxDir "binmac/basebox"
)

$BaseboxExec = $null
foreach ($candidate in $Candidates) {
    if (Test-Path $candidate) {
        $BaseboxExec = $candidate
        break
    }
}

if (-not $BaseboxExec) {
    Write-Error "Unable to locate the Basebox executable."
    exit 1
}

& $BaseboxExec -noprimaryconf -nolocalconf -conf $UserConfig @args
