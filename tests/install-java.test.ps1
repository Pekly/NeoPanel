# Offline Windows helper checks: no binary/runtime downloads or service changes.
$ErrorActionPreference = 'Stop'
$helper = Join-Path $PSScriptRoot '..\scripts\install-java.ps1'
$tokens = $null
$errors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile((Resolve-Path $helper), [ref]$tokens, [ref]$errors)
if ($errors.Count -ne 0) { throw "PowerShell syntax errors: $errors" }
$source = $ast.Extent.Text
foreach ($required in @('& $binary.FullName java install-all', '$LASTEXITCODE -ne 0', 'Get-Item -LiteralPath $BinaryPath', 'Push-Location -LiteralPath $binary.DirectoryName', 'Pop-Location')) {
    if (-not $source.Contains($required)) { throw "Missing helper contract: $required" }
}
$commands = $ast.FindAll({ param($node) $node -is [System.Management.Automation.Language.CommandAst] }, $true)
foreach ($command in $commands) {
    if ($command.GetCommandName() -match '^(Invoke-WebRequest|Invoke-RestMethod|Start-Process|New-Service|Start-Service|Set-Service|netsh|curl|wget)$') {
        throw "Unexpected setup side effect: $command"
    }
}
try {
    & $helper -BinaryPath (Join-Path $PSScriptRoot 'nonexistent-neopanel-java-fixture.exe')
    throw 'Missing binary was accepted'
} catch {
    if ($_.Exception.Message -eq 'Missing binary was accepted') { throw }
}
Write-Output 'PASS: Windows Java helper syntax, existing-binary contract, hosting directory, exit checking and no setup side effects'
