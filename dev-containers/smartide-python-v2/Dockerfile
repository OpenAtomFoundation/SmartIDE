#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM #{SMARTIDEBASEVERSION}#

USER root
RUN apt-get update && \
    apt-get -y install --no-install-recommends python2 python3 python3-distutils python3-pip python3-setuptools -y build-essential make && \
    apt-get clean && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*


USER smartide
# use sudo so that user does not get sudo usage info on (the first) login
RUN sudo echo "Running 'sudo' for smartide: success" && \
    # create .bashrc.d folder and source it in the bashrc
    mkdir -p /home/smartide/.bashrc.d && \
    (echo; echo "for i in \$(ls -A \$HOME/.bashrc.d/); do source \$HOME/.bashrc.d/\$i; done"; echo) >> /home/smartide/.bashrc

ENV HOME=/home/smartide
WORKDIR $HOME

ENV NODE_VERSION=14.17.6
ENV NVM_DIR /home/smartide/.nvm
RUN curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.38.0/install.sh | sh \
    && . $NVM_DIR/nvm.sh  \
    && nvm install $NODE_VERSION \
    && nvm install 12.22.7 \
    && nvm install 16.7.0 \
    && nvm alias default $NODE_VERSION \
    && nvm use $NODE_VERSION \
    && npm install -g yarn node-gyp \
    && echo ". ~/.nvm/nvm-lazy.sh"  >> /home/smartide/.bashrc.d/50-node
# above, we are adding the lazy nvm init to .bashrc, because one is executed on interactive shells, the other for non-interactive shells (e.g. plugin-host)
COPY --chown=smartide:smartide nvm-lazy.sh /home/smartide/.nvm/nvm-lazy.sh
ENV PATH $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH

USER root

RUN echo "export NVM_DIR_ROOT=\"/home/smartide/.nvm\"" >> /etc/profile && \
    echo "[ -s \"\$NVM_DIR_ROOT/nvm.sh\" ] && \. \"\$NVM_DIR_ROOT/nvm.sh\""  >> /etc/profile  && \
    echo "[ -s \"\$NVM_DIR_ROOT/bash_completion\" ] && \. \"\$NVM_DIR_ROOT/bash_completion\""  >> /etc/profile
