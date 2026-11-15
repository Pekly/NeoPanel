[CmdletBinding()]
param(
    [ValidateSet('amd64', 'arm64')][string]$Architecture = 'amd64',
    [string]$OutputDirectory,
    [switch]$NoArchive
)

$ErrorActionPreference = 'Stop'
$go = Get-Command go -ErrorAction Stop
$version = & $go.Source version
if ($LASTEXITCODE -ne 0 -or $version -notmatch 'go(\d+)\.(\d+)') { throw 'Could not detect the Go compiler version.' }
if ([int]$Matches[1] -lt 1 -or ([int]$Matches[1] -eq 1 -and [int]$Matches[2] -lt 26)) { throw 'Go 1.26 or newer is required.' }
if (!$OutputDirectory) { $OutputDirectory = Join-Path $PSScriptRoot "dist\windows-$Architecture" }
$OutputDirectory = [IO.Path]::GetFullPath($OutputDirectory)
$null = New-Item -ItemType Directory -Path $OutputDirectory -Force
$previous = @{ GOOS = $env:GOOS; GOARCH = $env:GOARCH; CGO_ENABLED = $env:CGO_ENABLED }
Push-Location $PSScriptRoot
try {
    $env:GOOS = 'windows'; $env:GOARCH = $Architecture; $env:CGO_ENABLED = '0'
    $panel = Join-Path $OutputDirectory 'NeoPanel.exe'
    $node = Join-Path $OutputDirectory 'NeoPanel-node.exe'
    & $go.Source build -trimpath '-ldflags=-s -w' -o $panel ./cmd/neopanel
    if ($LASTEXITCODE -ne 0) { throw 'Windows panel build failed.' }
    & $go.Source build -trimpath '-ldflags=-s -w' -tags node -o $node ./cmd/neopanel
    if ($LASTEXITCODE -ne 0) { throw 'Windows web-less node build failed.' }
    $launcher = Join-Path $OutputDirectory 'start-windows.ps1'
    $wrapper = Join-Path $OutputDirectory 'start-windows.cmd'
    $readme = Join-Path $OutputDirectory 'README-Windows.md'
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'start-windows.ps1') -Destination $launcher -Force
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'start-windows.cmd') -Destination $wrapper -Force
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'docs\windows.md') -Destination $readme -Force
    $files = @($panel, $node, $launcher, $wrapper, $readme)
    $sums = Join-Path $OutputDirectory 'SHA256SUMS.txt'
    $lines = $files | ForEach-Object { '{0}  {1}' -f (Get-FileHash -LiteralPath $_ -Algorithm SHA256).Hash.ToLowerInvariant(), [IO.Path]::GetFileName($_) }
    [IO.File]::WriteAllLines($sums, [string[]]$lines, [Text.Encoding]::ASCII)
    if (!$NoArchive) {
        # Package only known build files, never configuration, credentials, or server data.
        $archive = "$OutputDirectory.zip"
        Compress-Archive -LiteralPath ($files + $sums) -DestinationPath $archive -Force
        $hash = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
        [IO.File]::WriteAllText("$archive.sha256", "$hash  $([IO.Path]::GetFileName($archive))`r`n", [Text.Encoding]::ASCII)
        Write-Host "Package: $archive"
    }
    Write-Host "Windows $Architecture builds ready: $OutputDirectory"
    Write-Host 'These local builds are unsigned; code signing/release distribution is a separate step.'
} finally {
    $env:GOOS = $previous.GOOS; $env:GOARCH = $previous.GOARCH; $env:CGO_ENABLED = $previous.CGO_ENABLED
    Pop-Location
}
