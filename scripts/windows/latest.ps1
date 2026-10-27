# this script is used by the continuous integration server to
# build and publish the docker image for a commit to master.
$ErrorActionPreference = "Stop"

$env:GOOS="windows"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"

if (-not (Test-Path env:VERSION)) {
    $env:VERSION="1809"
}

if (-not (Test-Path env:REGISTRY)) {
    $env:REGISTRY="docker"
}

echo $env:GOOS
echo $env:GOARCH
echo $env:VERSION
echo $env:REGISTRY

# build the binary
. (Join-Path $PSScriptRoot "../setup-go-windows.ps1")
if ($env:REGISTRY -ne "docker") {
    go build -o release/windows/amd64/drone-docker.exe ./cmd/drone-docker
}
Write-Host "+ go build -o release/windows/amd64/drone-${env:REGISTRY}.exe ./cmd/drone-${env:REGISTRY}";
go build -o release/windows/amd64/drone-${env:REGISTRY}.exe ./cmd/drone-${env:REGISTRY}

# build and publish the docker image
docker login -u ${env:USERNAME} -p ${env:PASSWORD}
Write-Host "+ docker build --pull --no-cache -f docker/${env:REGISTRY}/Dockerfile.windows.amd64.${env:VERSION} -t plugins/${env:REGISTRY}:windows-${env:VERSION}-amd64 .";
docker build --pull --no-cache -f docker/${env:REGISTRY}/Dockerfile.windows.amd64.${env:VERSION} -t plugins/${env:REGISTRY}:windows-${env:VERSION}-amd64 .
Write-Host "+ docker push plugins/${env:REGISTRY}:windows-${env:VERSION}-amd64"
docker push plugins/${env:REGISTRY}:windows-${env:VERSION}-amd64

# remove images from local cache
Write-Host "+ docker rmi plugins/${env:REGISTRY}:windows-${env:VERSION}-amd64"
docker rmi plugins/${env:REGISTRY}:windows-${env:VERSION}-amd64
