#! /bin/bash
chown -R smartide:smartide /home/project


#获取传入的环境变量
export MARKETPLACE_URL=`cat /proc/1/environ | tr '\0' '\n' | grep 'MARKETPLACE_URL' | awk -F'=' '{print $2}'`


echo "Starting with MARKETPLACE_URL : $MARKETPLACE_URL"
marketplace=""
if [ ! -n "$MARKETPLACE_URL" ]; then
echo "-----ENV MARKETPLACE_URL -----IS NULL"
marketplace="https://marketplace.smartide.cn"
else
echo "-----ENV MARKETPLACE_URL -----NOT NULL"
marketplace=$MARKETPLACE_URL
fi
echo "-----marketplace:$marketplace"

cd /home/opvscode
find ./  -name "*.js" | xargs perl -pi -e "s|https://open-vsx.org|$marketplace|g" 
find ./  -name "*.json" | xargs perl -pi -e "s|https://open-vsx.org|$marketplace|g"

su smartide -c  "/home/smartide/.nvm/versions/node/v16.9.1/bin/node /home/opvscode/out/server-main.js --host 0.0.0.0 --without-connection-token"
