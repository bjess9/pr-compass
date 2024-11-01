# Detect the OS and Architecture
$OS = "windows"
$ARCH = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "arm64" }

# Set the download URL
if ($args.Length -gt 0) {
    $VERSION = $args[0]
} else {
    $VERSION = "latest"
}

try {
    if ($VERSION -eq "latest") {
        $latestRelease = Invoke-RestMethod https://api.github.com/repos/bjess9/pr-pilot/releases/latest
        $URL = ($latestRelease.assets | Where-Object { $_.name -like "*$OS*$ARCH*.tar.gz" }).browser_download_url
    } else {
        $URL = "https://github.com/bjess9/pr-pilot/releases/download/$VERSION/pr-pilot_${OS}_${ARCH}.tar.gz"
    }

    Write-Output "Download URL: $URL"
    
    if (-not $URL) {
        throw "No matching release found for $OS and $ARCH."
    }

    # Download and verify
    $DownloadPath = "$PSScriptRoot\pr-pilot.tar.gz"
    Invoke-WebRequest -Uri $URL -OutFile $DownloadPath

    # Check if download succeeded and is valid
    if (-not (Test-Path -Path $DownloadPath)) {
        throw "Download failed. Please check the URL or network connection."
    }

    # Ensure extract directory exists
    $ExtractPath = "$PSScriptRoot\pr-pilot"
    if (-not (Test-Path -Path $ExtractPath)) {
        New-Item -ItemType Directory -Path $ExtractPath | Out-Null
    }

    # Extract using tar (available on Windows 10+)
    tar -xzf $DownloadPath -C $ExtractPath

    # List contents of the extracted directory
    Write-Output "Contents of Extracted Directory:"
    Get-ChildItem -Path $ExtractPath -Recurse

    # Adjust path if pr-pilot.exe is nested
    $ExecutablePath = Get-ChildItem -Path $ExtractPath -Recurse -Filter "pr-pilot.exe" | Select-Object -First 1 -ExpandProperty FullName
    if (-not $ExecutablePath) {
        throw "Executable not found in extracted files."
    }

    # Ensure destination directory exists
    $DestinationPath = "C:\Program Files\pr-pilot"
    if (-not (Test-Path -Path $DestinationPath)) {
        New-Item -ItemType Directory -Path $DestinationPath | Out-Null
    }

    # Move to a directory in the system PATH
    Move-Item -Path $ExecutablePath -Destination "$DestinationPath\pr-pilot.exe" -Force

    # Add to system PATH if not already present
    $currentPath = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::Machine)
    if ($currentPath -notlike "*$DestinationPath*") {
        [System.Environment]::SetEnvironmentVariable("Path", $currentPath + ";$DestinationPath", [System.EnvironmentVariableTarget]::Machine)
        Write-Output "Added $DestinationPath to system PATH."
    } else {
        Write-Output "$DestinationPath is already in the system PATH."
    }

    # Clean up
    Remove-Item -Path $DownloadPath
    Remove-Item -Recurse -Force -Path $ExtractPath

    Write-Output "PR Pilot installed successfully! Run 'pr-pilot configure' to get started."
}
catch {
    Write-Output "Error: $_"
}
