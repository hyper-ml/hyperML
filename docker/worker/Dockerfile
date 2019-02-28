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



# Configure environment
ENV CONDA_DIR /opt/conda
ENV PATH $CONDA_DIR/bin:$PATH
ENV SHELL /bin/bash
ENV HF_USER hflow
ENV HF_UID 1000
ENV HOME /home/$HF_USER
ENV GO_HOME $HOME/go
ENV GO_PATH $HOME/go/bin
ENV PATH $GO_HOME/bin:$PATH
ENV LC_ALL en_US.UTF-8
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US.UTF-8

# Create user with UID=1000 and in the 'users' group
RUN useradd -m -s /bin/bash -N -u $HF_UID $HF_USER && \
 	mkdir -p $GO_HOME/bin && \
    mkdir -p $CONDA_DIR && \
    chown $HF_USER $CONDA_DIR 


USER $HF_USER


ENV MINICONDA_VERSION 4.3.21
RUN cd $HOME && \
    mkdir -p $CONDA_DIR && \
    wget --quiet https://repo.continuum.io/miniconda/Miniconda3-${MINICONDA_VERSION}-Linux-x86_64.sh && \
    echo "c1c15d3baba15bf50293ae963abef853 *Miniconda3-${MINICONDA_VERSION}-Linux-x86_64.sh" | md5sum -c - && \
    /bin/bash Miniconda3-${MINICONDA_VERSION}-Linux-x86_64.sh -f -b -p $CONDA_DIR && \
    rm Miniconda3-${MINICONDA_VERSION}-Linux-x86_64.sh && \
    conda clean -tipsy  
#     $CONDA_DIR/bin/conda config --system --prepend channels conda-forge && \
#    $CONDA_DIR/bin/conda config --system --set auto_update_conda false && \
#    $CONDA_DIR/bin/conda config --system --set show_channel_urls true && \
#    $CONDA_DIR/bin/conda update --all  

RUN conda install --quiet --yes  \
    'nomkl' \ 
    'pytorch' \
    'torchvision' && \
    conda clean -tipsy

RUN conda install --quiet --yes -c conda-forge \
    'keras' \
    'tensorflow' && \
    conda clean -tipsy


WORKDIR $GO_PATH
ADD ./workhorse . 

WORKDIR $HOME

# create conda default environment
# ADD ./conda.yml .
# RUN conda env create  --file conda.yml 


