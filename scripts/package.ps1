param(
    [string]$Configuration = "production"
)

$ErrorActionPreference = "Stop"

$projectRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$releaseRoot = Join-Path $projectRoot "release"
$appRoot = Join-Path $releaseRoot "KEVA"
$zipPath = Join-Path $projectRoot "KEVA.zip"
$frontendRoot = Join-Path $projectRoot "frontend"
$exeSource = Join-Path $projectRoot "build\bin\KEVA.exe"
$exeSourceWithoutExtension = Join-Path $projectRoot "build\bin\KEVA"

function Assert-InProject {
    param([string]$Path)

    $fullPath = [System.IO.Path]::GetFullPath($Path)
    $rootPath = [System.IO.Path]::GetFullPath($projectRoot)
    if (-not $fullPath.StartsWith($rootPath, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Refusing to operate outside project root: $fullPath"
    }
}

Assert-InProject $releaseRoot
Assert-InProject $zipPath

Push-Location $frontendRoot
try {
    npm.cmd run build
}
finally {
    Pop-Location
}

Push-Location $projectRoot
try {
    $env:GOCACHE = Join-Path $projectRoot ".gocache"
    wails build -clean -s -o KEVA.exe
}
finally {
    Pop-Location
}

if ((-not (Test-Path -LiteralPath $exeSource)) -and (Test-Path -LiteralPath $exeSourceWithoutExtension)) {
    $exeSource = $exeSourceWithoutExtension
}

if (-not (Test-Path -LiteralPath $exeSource)) {
    throw "Build output not found: $exeSource"
}

if (Test-Path -LiteralPath $releaseRoot) {
    Remove-Item -LiteralPath $releaseRoot -Recurse -Force
}
if (Test-Path -LiteralPath $zipPath) {
    Remove-Item -LiteralPath $zipPath -Force
}

New-Item -ItemType Directory -Force -Path $appRoot | Out-Null
foreach ($dir in @("config", "data", "logs", "resources", "resources\assets", "libs")) {
    New-Item -ItemType Directory -Force -Path (Join-Path $appRoot $dir) | Out-Null
}

Copy-Item -LiteralPath $exeSource -Destination (Join-Path $appRoot "KEVA.exe") -Force
Copy-Item -LiteralPath (Join-Path $projectRoot "build\windows\icon.ico") -Destination (Join-Path $appRoot "resources\icon.ico") -Force

@"
KEVA - KEy VAult

KEVA is a local-first Windows password manager. It stores vault data beside the app directory and does not require an installer, administrator rights, a user account, a cloud service, or registry writes.

Quick start
1. Extract KEVA.zip.
2. Double-click KEVA.exe.
3. Create a Master Password on first launch.
4. Add account records in the vault workspace.

Portable data
- Config is created at ./config/config.json after first initialization.
- The SQLite vault is created at ./data/keva.db.
- Logs, future assets, and runtime libraries stay under this KEVA folder.
- To move KEVA, copy the whole KEVA directory.
- To uninstall KEVA, delete the KEVA directory.

Security notes
- Do not forget your Master Password. KEVA does not store it and cannot recover it.
- Back up the whole KEVA folder, especially ./config and ./data, after creating real vault data.
- Account fields are encrypted before they are stored in SQLite.
- KEVA auto-locks after the configured idle period.

Windows integration
- System tray menu: Open KEVA, Lock, Exit.
- Global shortcut: Ctrl+Shift+K opens KEVA and focuses quick account search.
"@ | Set-Content -LiteralPath (Join-Path $appRoot "README.txt") -Encoding UTF8

Compress-Archive -LiteralPath $appRoot -DestinationPath $zipPath -Force

Write-Host "Created $zipPath"
