# Builds TakeoutRestore GUI + CLI.
# Full matrix (default): Windows/Linux/macOS amd64+arm64 where applicable.
# Local quick rebuild:  .\build.ps1 -Local
# Both GUI and CLI are always produced for every target in the chosen mode.
[CmdletBinding()]
param(
  [switch]$Local,
  [switch]$SkipVet,
  [string]$ExifToolVersion = "13.60"
)

$ErrorActionPreference = "Stop"
Set-Location -Path $PSScriptRoot

$GuiTags = "desktop,production,gui"
$Dist = Join-Path $PSScriptRoot "dist"

function Ensure-ExifTool {
  $zip = Join-Path $PSScriptRoot "assets\exiftool_win64.zip"
  if (Test-Path $zip) { return }
  Write-Host "Downloading ExifTool $ExifToolVersion for Windows..."
  New-Item -ItemType Directory -Force -Path (Join-Path $PSScriptRoot "assets") | Out-Null
  $url = "https://sourceforge.net/projects/exiftool/files/exiftool-${ExifToolVersion}_64.zip/download"
  Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing
}

function Ensure-Frontend {
  $index = Join-Path $PSScriptRoot "frontend\dist\index.html"
  $app = Join-Path $PSScriptRoot "frontend\dist\app.js"
  if (-not (Test-Path $index) -or -not (Test-Path $app)) {
    throw "frontend/dist is incomplete (need index.html and app.js for GUI builds)"
  }
}

function Ensure-WindowsIcon {
  $icon = Join-Path $PSScriptRoot "build\windows\icon256.png"
  if (-not (Test-Path $icon)) {
    Write-Host "Missing build/windows/icon256.png (app icon); continuing without refresh."
    return
  }
  $winres = Get-Command go-winres -ErrorAction SilentlyContinue
  if ($winres) {
    Write-Host "Refreshing Windows icon resource..."
    & go-winres simply `
      --icon $icon `
      --arch amd64 `
      --manifest gui `
      --product-name TakeoutRestore `
      --file-description "Restore Google Takeout photos" `
      --original-filename TakeoutRestore.exe `
      --product-version 0.1.0 `
      --file-version 0.1.0 `
      --copyright MIT `
      --out rsrc
    if ($LASTEXITCODE -ne 0) { throw "go-winres failed" }
  } elseif (-not (Test-Path (Join-Path $PSScriptRoot "rsrc_windows_amd64.syso"))) {
    Write-Warning "rsrc_windows_amd64.syso missing and go-winres not installed; exe will have no custom icon."
  }
}

function Copy-Docs([string]$DestDir) {
  Copy-Item docs\README.txt (Join-Path $DestDir "README.txt") -Force
  Copy-Item docs\README_ru.txt (Join-Path $DestDir "ПРОЧТИ.txt") -Force
}

function Invoke-GoBuild {
  param(
    [string]$Goos,
    [string]$Goarch,
    [string]$Out,
    [string]$Tags = "",
    [string]$Ldflags = "-s -w"
  )
  $env:GOOS = $Goos
  $env:GOARCH = $Goarch
  $env:CGO_ENABLED = "0"
  $args = @("build", "-trimpath", "-ldflags", $Ldflags, "-o", $Out, ".")
  if ($Tags) {
    $args = @("build", "-tags", $Tags, "-trimpath", "-ldflags", $Ldflags, "-o", $Out, ".")
  }
  Write-Host "go $($args -join ' ')  (GOOS=$Goos GOARCH=$Goarch)"
  & go @args
  if ($LASTEXITCODE -ne 0) { throw "go build failed for $Goos/$Goarch -> $Out" }
}

function Build-WindowsGui {
  Write-Host "Building Windows amd64 (desktop GUI)..."
  Ensure-WindowsIcon
  $exe = Join-Path $Dist "TakeoutRestore-windows-amd64-gui.exe"
  Invoke-GoBuild -Goos windows -Goarch amd64 -Out $exe -Tags $GuiTags -Ldflags "-s -w -H windowsgui"

  $pkgRoot = Join-Path $Dist "pkg"
  $pkgApp = Join-Path $pkgRoot "TakeoutRestore"
  Remove-Item $pkgRoot -Recurse -Force -ErrorAction SilentlyContinue
  New-Item -ItemType Directory -Force -Path $pkgApp | Out-Null
  # Zip keeps a short name for end users who double-click the app.
  Copy-Item $exe (Join-Path $pkgApp "TakeoutRestore.exe") -Force
  Copy-Docs $pkgApp

  $zipPath = Join-Path $Dist "TakeoutRestore.zip"
  if (Test-Path $zipPath) { Remove-Item $zipPath -Force }
  Compress-Archive -Path $pkgApp -DestinationPath $zipPath -Force
  Remove-Item $pkgRoot -Recurse -Force
}

function Build-UnixGui([string]$Os, [string]$Arch) {
  Write-Host "Building $Os $Arch (desktop GUI)..."
  $binName = "TakeoutRestore-$Os-$Arch"
  $binPath = Join-Path $Dist $binName
  Invoke-GoBuild -Goos $Os -Goarch $Arch -Out $binPath -Tags $GuiTags

  $stage = Join-Path $Dist ("stage-" + [guid]::NewGuid().ToString("n"))
  $appDir = Join-Path $stage "TakeoutRestore"
  New-Item -ItemType Directory -Force -Path $appDir | Out-Null
  Copy-Item $binPath (Join-Path $appDir "TakeoutRestore") -Force
  Copy-Docs $appDir

  $tarPath = Join-Path $Dist "$binName.tar.gz"
  if (Test-Path $tarPath) { Remove-Item $tarPath -Force }
  if (Get-Command tar -ErrorAction SilentlyContinue) {
    & tar -C $stage -czf $tarPath TakeoutRestore
    if ($LASTEXITCODE -ne 0) { throw "tar failed for $tarPath" }
  } else {
    Write-Warning "tar not found; left unpackaged binary at $binPath"
  }
  Remove-Item $stage -Recurse -Force
}

function Build-Cli([string]$Os, [string]$Arch) {
  Write-Host "Building $Os $Arch (CLI + web, no gui tag)..."
  $out = Join-Path $Dist "TakeoutRestore-$Os-$Arch-cli"
  if ($Os -eq "windows") { $out += ".exe" }
  Invoke-GoBuild -Goos $Os -Goarch $Arch -Out $out
}

Ensure-Frontend
Ensure-ExifTool
New-Item -ItemType Directory -Force -Path $Dist | Out-Null

if (-not $SkipVet) {
  Write-Host "go vet (CLI)..."
  & go vet ./...
  if ($LASTEXITCODE -ne 0) { throw "go vet failed" }
  Write-Host "go vet (GUI tags)..."
  & go vet -tags $GuiTags ./...
  if ($LASTEXITCODE -ne 0) { throw "go vet -tags $GuiTags failed" }
}

& go mod tidy
if ($LASTEXITCODE -ne 0) { throw "go mod tidy failed" }

# Clear cross-compile env when done
$oldGoos = $env:GOOS
$oldGoarch = $env:GOARCH
$oldCgo = $env:CGO_ENABLED
try {
  if ($Local) {
    Build-WindowsGui
    Build-Cli windows amd64
  } else {
    Build-WindowsGui
    Build-UnixGui linux amd64
    Build-UnixGui linux arm64
    Build-UnixGui darwin amd64
    Build-UnixGui darwin arm64

    Build-Cli windows amd64
    Build-Cli linux amd64
    Build-Cli linux arm64
    Build-Cli darwin amd64
    Build-Cli darwin arm64
  }
} finally {
  if ($null -eq $oldGoos) { Remove-Item Env:GOOS -ErrorAction SilentlyContinue } else { $env:GOOS = $oldGoos }
  if ($null -eq $oldGoarch) { Remove-Item Env:GOARCH -ErrorAction SilentlyContinue } else { $env:GOARCH = $oldGoarch }
  if ($null -eq $oldCgo) { Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue } else { $env:CGO_ENABLED = $oldCgo }
}

Write-Host ""
Write-Host "Artifacts in dist/:"
Get-ChildItem $Dist | Sort-Object Name | Format-Table Name, Length, LastWriteTime -AutoSize
