# Detect the OS and Architecture
$OS = "windows"
$ARCH = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

# Set the download URL
if ($args.Length -gt 0) {
    $VERSION = $args[0]
} else {
    $VERSION = "latest"
}

try {
    if ($VERSION -eq "latest") {
        $latestRelease = Invoke-RestMethod https://api.github.com/repos/bjess9/pr-pilot/releases/latest
        $URL = ($latestRelease.assets | Where-Object { $_.name -like "*$OS*$ARCH*" }).browser_download_url
    } else {
        $URL = "https://github.com/bjess9/pr-pilot/releases/download/$VERSION/pr-pilot_${OS}_${ARCH}.zip"
    }

    if (-not $URL) {
        throw "No matching release found for $OS and $ARCH."
    }

    # Download and verify
    $DownloadPath = "$PSScriptRoot\pr-pilot.zip"
    Invoke-WebRequest -Uri $URL -OutFile $DownloadPath

    # Check if download succeeded
    if (-not (Test-Path -Path $DownloadPath)) {
        throw "Download failed. Please check the URL or network connection."
    }

    # Unzip and move to a directory in the system PATH
    $ExtractPath = "$PSScriptRoot\pr-pilot"
    Expand-Archive -Path $DownloadPath -DestinationPath $ExtractPath -Force
    Move-Item -Path "$ExtractPath\pr-pilot.exe" -Destination "C:\Program Files\pr-pilot\pr-pilot.exe" -Force

    # Clean up
    Remove-Item -Path $DownloadPath
    Remove-Item -Recurse -Force -Path $ExtractPath

    Write-Output "PR Pilot installed successfully! Run 'pr-pilot configure' to get started."
}
catch {
    Write-Output "Error: $_"
}
