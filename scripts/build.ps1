# Define output directory
$OutputDir = "bin"
If (!(Test-Path -Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Define target platforms and architectures
$Platforms = @("windows", "linux", "darwin")
$Architectures = @("amd64", "386", "arm64", "arm")

# Build each combination
ForEach ($OS in $Platforms) {
    ForEach ($ARCH in $Architectures) {
        $OutputName = "fileorbit_${OS}_${ARCH}"

        # Skip windows/arm, darwin/arm and darwin/386
        If ($OS -eq "windows" -And $ARCH -eq "arm") {
            Continue
        } ElseIf ($OS -eq "darwin" -And $ARCH -eq "arm") {
            Continue
        } ElseIf ($OS -eq "darwin" -And $ARCH -eq "386") {
            Continue
        }

        # Windows binaries need .exe extension
        If ($OS -eq "windows") {
            $OutputName += ".exe"
        }
        
        Write-Host "Building for $OS/$ARCH..."
        
        # Set environment variables and build
        $env:CGO_ENABLED = "0"
        $env:GOOS = $OS
        $env:GOARCH = $ARCH
        go build -o "$OutputDir/$OutputName" .
        
        If ($?) {
            Write-Host "Successfully built: $OutputDir/$OutputName"
        } Else {
            Write-Host "Failed to build for $OS/$ARCH"
        }
    }
}

# Reset environment variables
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

Write-Host "All builds completed."
