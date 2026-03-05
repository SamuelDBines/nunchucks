$ErrorActionPreference = "Stop"

$repo = "SamuelDBines/nunchucks"
$version = if ($args.Length -gt 0) { $args[0] } else { "latest" }

function Get-InstallDir {
    if ($env:NUNCHUCKS_INSTALL_DIR) {
        return $env:NUNCHUCKS_INSTALL_DIR
    }

    $candidates = @(
        (Join-Path $HOME "bin"),
        (Join-Path $HOME ".local\bin")
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    return $candidates[0]
}

function Test-DirOnPath {
    param([string]$Dir)

    $parts = ($env:PATH -split ';') | Where-Object { $_ }
    foreach ($part in $parts) {
        if ([System.IO.Path]::GetFullPath($part) -eq [System.IO.Path]::GetFullPath($Dir)) {
            return $true
        }
    }

    return $false
}

function Add-UserPath {
    param([string]$Dir)

    $current = [Environment]::GetEnvironmentVariable("Path", "User")
    $parts = @()
    if ($current) {
        $parts = $current -split ';' | Where-Object { $_ }
    }

    if ($parts -contains $Dir) {
        return
    }

    $updated = @($parts + $Dir) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $updated, "User")
    $env:PATH = if ($env:PATH) { "$env:PATH;$Dir" } else { $Dir }
}

$installDir = Get-InstallDir

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "Unsupported Windows architecture: $($env:PROCESSOR_ARCHITECTURE)" }
}

if ($version -eq "latest") {
    $release = Invoke-WebRequest -Uri "https://api.github.com/repos/$repo/releases/latest" | ConvertFrom-Json
    $version = $release.tag_name
}

$asset = "nunchucks_${version}_windows_${arch}.zip"
$url = "https://github.com/$repo/releases/download/$version/$asset"
$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())

New-Item -ItemType Directory -Path $tmpDir | Out-Null
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$archive = Join-Path $tmpDir $asset
Invoke-WebRequest -Uri $url -OutFile $archive
Expand-Archive -Path $archive -DestinationPath $tmpDir -Force
Copy-Item -Path (Join-Path $tmpDir "nunchucks_${version}_windows_${arch}\nunchucks.exe") -Destination (Join-Path $installDir "nunchucks.exe") -Force

$binary = Join-Path $installDir "nunchucks.exe"

if (-not (Test-DirOnPath -Dir $installDir)) {
    Add-UserPath -Dir $installDir
    Write-Host "added $installDir to your user PATH"
}

& $binary version | Out-Null

Write-Host "installed nunchucks $version to $binary"
Write-Host "run: nunchucks.exe help"
