$ErrorActionPreference = "Stop"

$repo = "SamuelDBines/nunjucks"
$version = if ($args.Length -gt 0) { $args[0] } else { "latest" }
$installDir = if ($env:NUNCHUCKS_INSTALL_DIR) { $env:NUNCHUCKS_INSTALL_DIR } else { Join-Path $HOME "bin" }

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

Write-Host "installed nunchucks $version to $(Join-Path $installDir 'nunchucks.exe')"
