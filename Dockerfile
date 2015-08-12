# Docker image for the docker plugin
#
#     docker build --rm=true -t plugins/drone-docker .

FROM ubuntu:14.04

RUN apt-get update -qq

RUN apt-get --force-yes -y install curl                        
RUN apt-get --force-yes -y install apt-transport-https                       
RUN apt-get --force-yes -y install ca-certificates                              
RUN apt-get --force-yes -y install curl   
RUN apt-get --force-yes -y install lxc     
RUN apt-get --force-yes -y install iptables 
RUN curl -sSL https://get.docker.com/ubuntu/ -o /tmp/install.sh  
RUN sed -i -e "s/apt-get install/apt-get install --force-yes -y/g" /tmp/install.sh
RUN sh /tmp/install.sh
RUN rm -rf /var/lib/apt/lists/*

ADD drone-docker /go/bin/
ADD wrapdocker /bin/
RUN chmod u+x /bin/wrapdocker

VOLUME /var/lib/docker
ENTRYPOINT ["/go/bin/drone-docker"]
