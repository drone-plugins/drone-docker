# escape=`
FROM mcr.microsoft.com/windows/servercore:1809 as download

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]

ENV DOCKER_VERSION 18.09.1

RUN Invoke-WebRequest 'http://constexpr.org/innoextract/files/innoextract-1.6-windows.zip' -OutFile 'innoextract.zip' -UseBasicParsing ; `
    Expand-Archive innoextract.zip -DestinationPath C:\ ; `
    Remove-Item -Path innoextract.zip

RUN [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 ; `
    Invoke-WebRequest $('https://github.com/docker/toolbox/releases/download/v{0}/DockerToolbox-{0}.exe' -f $env:DOCKER_VERSION) -OutFile 'dockertoolbox.exe' -UseBasicParsing
RUN /innoextract.exe dockertoolbox.exe

FROM plugins/base:windows-1809

LABEL maintainer="Drone.IO Community <drone-dev@googlegroups.com>" `
  org.label-schema.name="Drone Docker" `
  org.label-schema.vendor="Drone.IO Community" `
  org.label-schema.schema-version="1.0"

COPY --from=download /windows/system32/netapi32.dll /windows/system32/netapi32.dll
COPY --from=download /app/docker.exe C:/bin/docker.exe
ADD release/windows/amd64/drone-docker.exe C:/bin/drone-docker.exe
ENTRYPOINT [ "C:\\bin\\drone-docker.exe" ]
