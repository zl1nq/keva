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
$rootIcon = Join-Path $projectRoot "icon.ico"
$rootPngIcon = Join-Path $projectRoot "appicon.png"
$wailsWindowsIcon = Join-Path $projectRoot "build\windows\icon.ico"
$wailsAppIcon = Join-Path $projectRoot "build\appicon.png"
$binResources = Join-Path $projectRoot "build\bin\resources"
$binAssets = Join-Path $binResources "assets"

function Assert-InProject {
    param([string]$Path)

    $fullPath = [System.IO.Path]::GetFullPath($Path)
    $rootPath = [System.IO.Path]::GetFullPath($projectRoot)
    if (-not $fullPath.StartsWith($rootPath, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Refusing to operate outside project root: $fullPath"
    }
}

function Assert-NoVaultData {
    param([string]$Path)

    if (-not (Test-Path -LiteralPath $Path)) {
        return
    }

    $configPath = Join-Path $Path "config\config.json"
    $databasePath = Join-Path $Path "data\keva.db"
    if ((Test-Path -LiteralPath $configPath) -or (Test-Path -LiteralPath $databasePath)) {
        throw "Refusing to delete release directory because vault data exists under: $Path"
    }
}

Assert-InProject $releaseRoot
Assert-InProject $zipPath

if (Test-Path -LiteralPath $rootIcon) {
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $wailsWindowsIcon) | Out-Null
    Copy-Item -LiteralPath $rootIcon -Destination $wailsWindowsIcon -Force
}
if (Test-Path -LiteralPath $rootPngIcon) {
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $wailsAppIcon) | Out-Null
    Copy-Item -LiteralPath $rootPngIcon -Destination $wailsAppIcon -Force
}

Push-Location $frontendRoot
try {
    npm.cmd run build
    if ($LASTEXITCODE -ne 0) {
        throw "Frontend build failed with exit code $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}

Push-Location $projectRoot
try {
    $env:GOCACHE = Join-Path $projectRoot ".gocache"
    wails build -s -o KEVA.exe
    if ($LASTEXITCODE -ne 0) {
        throw "Wails build failed with exit code $LASTEXITCODE"
    }
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

New-Item -ItemType Directory -Force -Path $binResources | Out-Null
New-Item -ItemType Directory -Force -Path $binAssets | Out-Null
if (Test-Path -LiteralPath $wailsWindowsIcon) {
    Copy-Item -LiteralPath $wailsWindowsIcon -Destination (Join-Path $binResources "icon.ico") -Force
}
if (Test-Path -LiteralPath $wailsAppIcon) {
    Copy-Item -LiteralPath $wailsAppIcon -Destination (Join-Path $binAssets "appicon.png") -Force
}

if (Test-Path -LiteralPath $appRoot) {
    Assert-NoVaultData $appRoot
    Remove-Item -LiteralPath $appRoot -Recurse -Force
}
if (Test-Path -LiteralPath $zipPath) {
    Remove-Item -LiteralPath $zipPath -Force
}

New-Item -ItemType Directory -Force -Path $appRoot | Out-Null
foreach ($dir in @("config", "data", "logs", "resources", "resources\assets", "libs")) {
    New-Item -ItemType Directory -Force -Path (Join-Path $appRoot $dir) | Out-Null
}

Copy-Item -LiteralPath $exeSource -Destination (Join-Path $appRoot "KEVA.exe") -Force
Copy-Item -LiteralPath $wailsWindowsIcon -Destination (Join-Path $appRoot "resources\icon.ico") -Force
if (Test-Path -LiteralPath $wailsAppIcon) {
    Copy-Item -LiteralPath $wailsAppIcon -Destination (Join-Path $appRoot "resources\assets\appicon.png") -Force
}

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
