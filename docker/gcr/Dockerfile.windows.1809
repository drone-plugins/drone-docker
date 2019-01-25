# escape=`
FROM plugins/docker:windows-1809

LABEL maintainer="Drone.IO Community <drone-dev@googlegroups.com>" `
  org.label-schema.name="Drone GCR" `
  org.label-schema.vendor="Drone.IO Community" `
  org.label-schema.schema-version="1.0"

ADD release/windows/amd64/drone-gcr.exe C:/bin/drone-gcr.exe
ENTRYPOINT [ "C:\\bin\\drone-gcr.exe" ]
