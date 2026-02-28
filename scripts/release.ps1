# Check if GitHub CLI is installed
if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    Write-Host "GitHub CLI (gh) is not installed. Please install it first."
    exit 1
}

# Check if version.txt exists
if (-not (Test-Path "version.txt")) {
    Write-Host "version.txt not found. Please create it with the release version."
    exit 1
}

# Read the tag from version.txt
$tag = Get-Content "version.txt" | ForEach-Object { $_.Trim() }
$releaseName = "Release $tag"

Write-Host "Creating GitHub release for tag: $tag"

# Get all files in the 'bin' directory
$files = Get-ChildItem -Path "bin" -File

# Ensure that we have at least one file
if ($files.Count -eq 0) {
    Write-Host "No files found in the 'bin' directory."
    exit 1
}

# Build file path list for gh
$filePaths = @($files | ForEach-Object { $_.FullName })

# Create the release with the files
gh release create "$tag" @filePaths --title "$releaseName" --notes "Automated release for $tag"

if ($?) {
    Write-Host "GitHub release created successfully."
} else {
    Write-Host "Failed to create GitHub release."
    exit 1
}
