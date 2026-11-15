[CmdletBinding()]
param(
    [ValidateSet('amd64', 'arm64')][string]$Architecture = 'amd64',
    [string]$OutputDirectory,
    [switch]$Installer,
    [string]$InnoCompiler
)
$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'Build the desktop package on Windows.' }
$go = (Get-Command go -ErrorAction Stop).Source
$version = & $go version
if ($LASTEXITCODE -ne 0 -or $version -notmatch 'go(\d+)\.(\d+)' -or [int]$Matches[1] -lt 1 -or ([int]$Matches[1] -eq 1 -and [int]$Matches[2] -lt 26)) { throw 'Go 1.26 or newer is required.' }
if (!$OutputDirectory) { $OutputDirectory = Join-Path $PSScriptRoot "dist\windows-desktop-$Architecture" }
$OutputDirectory = [IO.Path]::GetFullPath($OutputDirectory)
$null = New-Item -ItemType Directory -Path $OutputDirectory -Force
$resourcePrefix = Join-Path $PSScriptRoot 'cmd\neopanel-desktop\desktop_resources'
$resource = "${resourcePrefix}_windows_${Architecture}.syso"
if (Test-Path -LiteralPath $resource) { throw "Refusing to overwrite existing generated resource: $resource" }
$tempRoot = if ($env:PI_SCRATCH_DIR) { $env:PI_SCRATCH_DIR } else { [IO.Path]::GetTempPath() }
$temp = Join-Path $tempRoot ('neopanel-desktop-build-' + [Guid]::NewGuid().ToString('N'))
$null = New-Item -ItemType Directory -Path $temp
$previous = @{ GOOS = $env:GOOS; GOARCH = $env:GOARCH; CGO_ENABLED = $env:CGO_ENABLED }
Push-Location $PSScriptRoot
try {
    # Render NeoPanel's existing three-layer mark with Windows' code-native drawing API.
    # All intermediates live in the temporary build directory, never beside user data.
    Add-Type -AssemblyName System.Drawing
    $image = New-Object Drawing.Bitmap 256,256
    $graphics = [Drawing.Graphics]::FromImage($image)
    $pen = New-Object Drawing.Pen ([Drawing.Color]::FromArgb(16,185,129)),18
    try {
        $graphics.SmoothingMode = [Drawing.Drawing2D.SmoothingMode]::AntiAlias
        $graphics.Clear([Drawing.Color]::FromArgb(17,21,28))
        $pen.StartCap = [Drawing.Drawing2D.LineCap]::Round
        $pen.EndCap = [Drawing.Drawing2D.LineCap]::Round
        $pen.LineJoin = [Drawing.Drawing2D.LineJoin]::Round
        $graphics.DrawPolygon($pen, [Drawing.PointF[]]@([Drawing.PointF]::new(128,32),[Drawing.PointF]::new(32,80),[Drawing.PointF]::new(128,128),[Drawing.PointF]::new(224,80)))
        $graphics.DrawLines($pen, [Drawing.PointF[]]@([Drawing.PointF]::new(32,128),[Drawing.PointF]::new(128,176),[Drawing.PointF]::new(224,128)))
        $graphics.DrawLines($pen, [Drawing.PointF[]]@([Drawing.PointF]::new(32,176),[Drawing.PointF]::new(128,224),[Drawing.PointF]::new(224,176)))
        $image.Save((Join-Path $temp 'appicon.png'), [Drawing.Imaging.ImageFormat]::Png)
    } finally { $pen.Dispose(); $graphics.Dispose(); $image.Dispose() }
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'build\windows-desktop\winres.json') -Destination (Join-Path $temp 'winres.json')
    # Run build tools for the current host, even when the target executable is ARM64.
    $env:GOOS = 'windows'; $env:GOARCH = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }; $env:CGO_ENABLED = '0'
    $savedPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = 'Continue'
        & $go run github.com/tc-hib/go-winres@v0.3.3 make --in (Join-Path $temp 'winres.json') --out $resourcePrefix --arch $Architecture
        $code = $LASTEXITCODE
    } finally { $ErrorActionPreference = $savedPreference }
    if ($code -ne 0 -or !(Test-Path -LiteralPath $resource)) { throw 'Windows resource generation failed.' }
    $env:GOARCH = $Architecture
    $exe = Join-Path $OutputDirectory 'NeoPanel-Desktop.exe'
    & $go build -trimpath -tags 'desktop,production' '-ldflags=-s -w -H windowsgui' -o $exe ./cmd/neopanel-desktop
    if ($LASTEXITCODE -ne 0) { throw 'Desktop build failed. Close the app if replacing an existing EXE.' }
    $readme = Join-Path $OutputDirectory 'README-Desktop.md'
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'docs\windows-desktop.md') -Destination $readme -Force
    $files = @($exe, $readme)
    if ($Installer) {
        if (!$InnoCompiler) { $cmd = Get-Command ISCC.exe -ErrorAction SilentlyContinue; if ($cmd) { $InnoCompiler = $cmd.Source } }
        if (!$InnoCompiler) { $InnoCompiler = Join-Path ${env:ProgramFiles(x86)} 'Inno Setup 6\ISCC.exe' }
        if (!(Test-Path -LiteralPath $InnoCompiler -PathType Leaf)) { throw 'Install Inno Setup 6 and pass -InnoCompiler to build the optional setup EXE. The portable desktop EXE is already built.' }
        & $InnoCompiler "/DSourceDir=$OutputDirectory" "/DTargetArch=$Architecture" "/O$OutputDirectory" (Join-Path $PSScriptRoot 'build\windows-desktop\installer.iss')
        if ($LASTEXITCODE -ne 0) { throw 'Installer compilation failed.' }
        $files += Join-Path $OutputDirectory 'NeoPanel-Desktop-Setup.exe'
    }
    $sums = Join-Path $OutputDirectory 'SHA256SUMS.txt'
    [IO.File]::WriteAllLines($sums, [string[]]@($files | ForEach-Object { '{0}  {1}' -f (Get-FileHash -LiteralPath $_ -Algorithm SHA256).Hash.ToLowerInvariant(), [IO.Path]::GetFileName($_) }), [Text.Encoding]::ASCII)
    Compress-Archive -LiteralPath ($files + $sums) -DestinationPath "$OutputDirectory.zip" -Force
    $hash = (Get-FileHash -LiteralPath "$OutputDirectory.zip" -Algorithm SHA256).Hash.ToLowerInvariant()
    [IO.File]::WriteAllText("$OutputDirectory.zip.sha256", "$hash  $([IO.Path]::GetFileName($OutputDirectory)).zip`r`n", [Text.Encoding]::ASCII)
    Write-Host "Desktop app ready: $exe"
    Write-Host 'Double-click the EXE. No console, browser tab, Go installation, or administrator rights needed to run it.'
    Write-Host 'Requires Microsoft Edge WebView2 Runtime. Local builds are unsigned.'
} finally {
    if (Test-Path -LiteralPath $resource) { Remove-Item -LiteralPath $resource -Force }
    Remove-Item -LiteralPath $temp -Recurse -Force
    $env:GOOS = $previous.GOOS; $env:GOARCH = $previous.GOARCH; $env:CGO_ENABLED = $previous.CGO_ENABLED
    Pop-Location
}
