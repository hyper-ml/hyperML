FROM debian@sha256:52af198afd8c264f1035206ca66a5c48e602afb32dc912ebf9e9478134601ec4

MAINTAINER hyperML <beekeepr@hyperml.com>

USER root

ENV DEBIAN_FRONTEND noninteractive
RUN REPO=http://cdn-fastly.deb.debian.org \
 && echo "deb $REPO/debian jessie main\ndeb $REPO/debian-security jessie/updates main" > /etc/apt/sources.list \
 && apt-get update && apt-get -yq dist-upgrade \
 && apt-get install -yq --no-install-recommends \
    wget \
    bzip2 \
    ca-certificates \
    sudo \
    locales \
    fonts-liberation \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

RUN echo "en_US.UTF-8 UTF-8" > /etc/locale.gen && \
    locale-gen


RUN apt-get update && apt-get install -yq --no-install-recommends \
	build-essential \
	git \
	unzip \
	&& apt-get clean && \
    rm -rf /var/lib/apt/lists/*

 
# Configure environment vars
ENV SHELL /bin/bash
ENV HF_USER hflow
ENV HF_UID 1000
ENV HOME /home/$HF_USER
ENV HFL_HOME $HOME/.hflow
ENV GO_HOME $HOME/go
ENV GO_PATH $HOME/go/bin
ENV PATH $GO_HOME/bin:$PATH
ENV LC_ALL en_US.UTF-8
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US.UTF-8

# Create user with UID=1000 and in the 'users' group
RUN useradd -m -s /bin/bash -N -u $HF_UID $HF_USER && \
 	mkdir -p $GO_HOME/bin && \
    mkdir -p $HOME/.hflow 

# USER $HF_USER

WORKDIR $GO_PATH
ADD ./hflserver . 

# WORKDIR $HOME
ADD ./server_config .hflserver
ADD ./gcloud_config.json .hflow/gcs.json

COPY --from=lachlanevenson/k8s-kubectl:v1.10.3 /usr/local/bin/kubectl /usr/local/bin/kubectl

CMD ["hflserver"]
