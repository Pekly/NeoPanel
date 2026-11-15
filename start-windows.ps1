[CmdletBinding()]
param(
    [ValidateSet('panel', 'node')][string]$Role = 'panel',
    [string]$ConfigPath,
    [string]$DataDirectory,
    [string]$ListenHost,
    [ValidateRange(1, 65535)][int]$Port = 8080,
    [string]$BinaryPath,
    [switch]$InitializeOnly
)

$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'Use this launcher on Windows. Linux nodes use the normal binary/install.sh.' }
$exeName = if ($Role -eq 'node') { 'NeoPanel-node.exe' } else { 'NeoPanel.exe' }
if (!$BinaryPath) {
    $BinaryPath = Join-Path $PSScriptRoot $exeName
    if (!(Test-Path -LiteralPath $BinaryPath -PathType Leaf)) {
        $arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
        $BinaryPath = Join-Path $PSScriptRoot "dist\windows-$arch\$exeName"
    }
}
if (!(Test-Path -LiteralPath $BinaryPath -PathType Leaf)) {
    throw "Missing $exeName. Run build-windows.ps1 first, or use the extracted Windows package."
}
$BinaryPath = (Resolve-Path -LiteralPath $BinaryPath).Path
$arguments = @('-role', $Role)
if ($ConfigPath) { $arguments += @('-config', [IO.Path]::GetFullPath($ConfigPath)) }
if ($PSBoundParameters.ContainsKey('ListenHost')) { $arguments += @('-host', $ListenHost) }
if ($PSBoundParameters.ContainsKey('Port')) { $arguments += @('-port', [string]$Port) }
if ($InitializeOnly) { $arguments += '-init' }
$previousDataDir = $env:NEOPANEL_DATA_DIR
try {
    if ($DataDirectory) { $env:NEOPANEL_DATA_DIR = [IO.Path]::GetFullPath($DataDirectory) }
    Write-Host "Starting NeoPanel $Role. Default HTTP binding is localhost; use the URL printed below."
    Write-Host 'This is a foreground console process. Use Ctrl+C for graceful shutdown; do not close/kill it to stop a world.'
    Write-Host 'Local Windows workloads are trusted-only: Job Objects are not filesystem/network isolation.'
    # Native tools log to stderr even on success. Judge failure by exit code,
    # not PowerShell's NativeCommandError conversion when output is captured.
    $savedPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = 'Continue'
        & $BinaryPath @arguments
        $exitCode = $LASTEXITCODE
    } finally { $ErrorActionPreference = $savedPreference }
    if ($exitCode -ne 0) { throw "NeoPanel exited with code $exitCode" }
} finally {
    $env:NEOPANEL_DATA_DIR = $previousDataDir
}
