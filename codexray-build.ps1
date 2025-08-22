$targets = @(
  "linux amd64",
  "linux arm64",
  "windows amd64",
  "windows arm64",
  "darwin amd64",
  "darwin arm64"
)

foreach ($target in $targets) {
  $parts = $target -split " "
  $os = $parts[0]
  $arch = $parts[1]
  $ext = ""
  if ($os -eq "windows") { $ext = ".exe" }
  $output = "codexray-transformer-$os-$arch$ext"
  Write-Host "Building $output"
  $env:GOOS=$os
  $env:GOARCH=$arch
  go build -o $output .
}
