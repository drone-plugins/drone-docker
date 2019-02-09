# escape=`
FROM plugins/docker:windows-1809

LABEL maintainer="Drone.IO Community <drone-dev@googlegroups.com>" `
  org.label-schema.name="Drone ECR" `
  org.label-schema.vendor="Drone.IO Community" `
  org.label-schema.schema-version="1.0"

ADD release/windows/amd64/drone-ecr.exe C:/bin/drone-ecr.exe
ENTRYPOINT [ "C:\\bin\\drone-ecr.exe" ]
