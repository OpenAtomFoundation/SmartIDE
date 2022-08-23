#! /bin/bash
chown -R smartide:smartide /home/project
su smartide -c  "/home/smartide/.nvm/versions/node/v16.9.1/bin/node /home/opvscode/out/server-main.js --host 0.0.0.0 --without-connection-token"
