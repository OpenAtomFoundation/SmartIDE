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

echo 'ide.sh............start'

if [ -d "./openvscode-images/" ];then
 sudo rm -rf openvscode-images
else
  echo 'openvscode-server............不存在'
fi

if [ -d "./vsix/" ];then
 sudo rm -rf vsix
else
  echo 'vsix...........不存在'
fi

sudo mkdir openvscode-images vsix vsix/extensions
sudo chmod -R 777 openvscode-images
sudo chmod -R 777 vsix
sudo chmod -R 777 vsix/extensions


# 解压目录
sudo tar -zxf #{OpenVScodeServerFileName}#.tar.gz --strip-components 1 -C openvscode-images

# 删除node   
sudo rm -rf ./openvscode-images/node

# 删除server.sh
# sudo rm -rf ./openvscode-images/server.sh
# 复制server.sh
# sudo cp server.sh ./openvscode-images/
# sudo chmod +x ./openvscode-images/server.sh;


# 解压插件
OPVSCODEVSIX=./vsix

for i in ./extensions/*.vsix;
    do
    sudo unzip $i "extension/*" -d $OPVSCODEVSIX/extensions/$(basename -s .vsix $i); \
    sudo mv $OPVSCODEVSIX/extensions/$(basename -s .vsix $i)/extension/* $OPVSCODEVSIX/extensions/$(basename -s .vsix $i); \
    sudo rm -rf $OPVSCODEVSIX/extensions/$(basename -s .vsix $i)/extension; \
    echo "$i........已复制"; \
    done

sudo \cp -rf ./vsix/extensions openvscode-images

echo 'ide.sh............end'

