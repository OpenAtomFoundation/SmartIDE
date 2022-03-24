#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:2782

USER root

# Install Apache2 PHP
RUN apt-get update \
    && apt-get install -y software-properties-common ca-certificates \
    && add-apt-repository -y ppa:ondrej/php \
    && apt-get install -y \
        apache2 \
        php7.4 php7.4-xml php7.4-mysql php7.4-json php7.4-mbstring php7.4-ldap php7.4-soap php7.4-zip php7.4-gd php7.4-curl php7.4-xdebug php-apcu \
        graphviz \
        curl \
        unzip\
        git \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && update-alternatives --set php /usr/bin/php7.4

# Copy Apache services, configs and scripts
COPY artifacts/apache2.conf /etc/apache2/apache2.conf
COPY artifacts/apache2.fqdn.conf /etc/apache2/conf-available/fqdn.conf
COPY artifacts/apache2.sites-available.conf /etc/apache2/sites-available/000-default.conf
COPY artifacts/scripts /
RUN chmod +x /*.sh \
    && a2enconf fqdn

#clean
RUN apt-get clean && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*