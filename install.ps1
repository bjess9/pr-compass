# Detect the OS and Architecture
$OS = "windows"
$ARCH = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

# Set the download URL
if ($args.Length -gt 0) {
    $VERSION = $args[0]
} else {
    $VERSION = "latest"
}

if ($VERSION -eq "latest") {
    $URL = (Invoke-RestMethod https://api.github.com/repos/bjess9/pr-pilot/releases/latest).assets |
           Where-Object { $_.name -like "*$OS*$ARCH*" } |
           Select-Object -First 1 -ExpandProperty browser_download_url
} else {
    $URL = "https://github.com/bjess9/pr-pilot/releases/download/$VERSION/pr-pilot_${OS}_${ARCH}.zip"
}

# Download and extract
$DownloadPath = "$PSScriptRoot\pr-pilot.zip"
Invoke-WebRequest -Uri $URL -OutFile $DownloadPath

# Unzip and move to a directory in the system PATH
$ExtractPath = "$PSScriptRoot\pr-pilot"
Expand-Archive -Path $DownloadPath -DestinationPath $ExtractPath -Force
Move-Item -Path "$ExtractPath\pr-pilot.exe" -Destination "C:\Program Files\pr-pilot\pr-pilot.exe" -Force

# Clean up
Remove-Item -Path $DownloadPath
Remove-Item -Recurse -Force -Path $ExtractPath

Write-Output "PR Pilot installed successfully! Run 'pr-pilot configure' to get started."
