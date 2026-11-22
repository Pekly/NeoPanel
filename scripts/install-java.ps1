# Run as the SAME Windows account that hosts NeoPanel (no elevation required).
# Uses that account's managed runtime directory, or its absolute NEOPANEL_JAVA_DIR.
# No binary download, service registration, firewall changes or application startup.
[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$BinaryPath
)
$ErrorActionPreference = 'Stop'
$binary = Get-Item -LiteralPath $BinaryPath -ErrorAction Stop
if ($binary.PSIsContainer -or $binary.Extension -ne '.exe') {
    throw 'BinaryPath must name an existing NeoPanel backend executable (.exe).'
}
# Keep invocation in the hosting binary directory, independent of caller location.
Push-Location -LiteralPath $binary.DirectoryName
try {
    & $binary.FullName java install-all
    if ($LASTEXITCODE -ne 0) {
        throw "Java provisioning incomplete (exit $LASTEXITCODE). Resolve the reported missing family and rerun under the hosting account."
    }
} finally {
    Pop-Location
}
