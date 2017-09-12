FROM hyperhq/docker-in-hyper
MAINTAINER Amos (LFlare) Ng <amosng1@gmail.com>

# Update packages
RUN yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo && \
    yum -y remove docker* && \
    yum -y install docker-ce

# Add entrypoint
ADD ./dockerd-entrypoint.sh /usr/local/bin/dockerd-entrypoint.sh

# Add drone-docker binary
ADD release/linux/amd64/drone-docker /bin/

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-docker"]
