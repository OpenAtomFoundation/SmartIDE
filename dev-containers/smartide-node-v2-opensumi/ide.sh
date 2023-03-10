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

if [ -d "./opensumi-release/" ];then
 sudo rm -rf opensumi-release
else
  echo 'opensumi-release............不存在'
fi

if [ -d "./opensumi-extension/" ];then
 sudo rm -rf opensumi-extension
else
  echo 'opensumi-extension...........不存在'
fi

sudo mkdir opensumi-release opensumi-extension
sudo chmod -R 777 opensumi-release
sudo chmod -R 777 opensumi-extension


# 解压opensumi-release
sudo tar -zxf opensumi-release.tar.gz --strip-components 1 -C opensumi-release

# 解压opensumi-extension
sudo tar -zxf opensumi-extension.tar.gz --strip-components 1 -C opensumi-extension


echo 'ide.sh............end'