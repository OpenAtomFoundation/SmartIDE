#! /bin/bash
###########################################################################
# SmartIDE - Dev Containers
# Copyright (C) 2023 leansoftX.com

# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
###########################################################################

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
