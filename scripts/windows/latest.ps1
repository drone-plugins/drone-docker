# this script is used by the continuous integration server to
# build and publish the docker image for a commit to master.

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
Write-Host "+ go build -o release/windows/amd64/drone-$env:REGISTRY.exe";
go build -o release/windows/amd64/drone-$env:REGISTRY.exe

# build and publish the docker image
docker login -u $env:USERNAME -p $env:PASSWORD
Write-Host "+ docker build -f docker/$env:REGISTRY/Dockerfile.windows.amd64.$env:VERSION -t plugins/$env:REGISTRY:windows-$env:VERSION-amd64 .";
docker build -f docker/$env:REGISTRY/Dockerfile.windows.amd64.$env:VERSION -t plugins/$env:REGISTRY:windows-$env:VERSION-amd64 .
Write-Host "+ docker push plugins/$env:REGISTRY:windows-$env:VERSION-amd64"
docker push plugins/$env:REGISTRY:windows-$env:VERSION-amd64

# remove images from local cache
Write-Host "+ docker rmi plugins/$env:REGISTRY:windows-$env:VERSION-amd64"
docker rmi plugins/$env:REGISTRY:windows-$env:VERSION-amd64
