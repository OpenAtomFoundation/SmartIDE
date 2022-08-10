#! /bin/bash
chown -R smartide:smartide /home/project

export JAVA_HOME=/usr/lib/jvm/java-1.17.0-openjdk-amd64
export M2_HOME=/opt/maven
export MAVEN_HOME=/opt/maven
export PATH=${M2_HOME}/bin:${PATH}

su smartide -c  "/home/smartide/.nvm/versions/node/v16.9.1/bin/node /home/opvscode/out/server-main.js --host 0.0.0.0 --without-connection-token"
