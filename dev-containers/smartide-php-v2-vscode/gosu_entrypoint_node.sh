#!/bin/bash
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

USER_UID=${LOCAL_USER_UID:-1000}
USER_GID=${LOCAL_USER_GID:-1000}
USER_PASS=${LOCAL_USER_PASSWORD:-"smartide123.@IDE"}
DISABLE_CHOWN=${DISABLE_CHOWN:-0}
USERNAME=smartide

echo "gosu_entrypoint_node.sh"
echo "Starting with USER_UID : $USER_UID"
echo "Starting with USER_GID : $USER_GID"
echo "Starting with USER_PASS : $USER_PASS"
echo "Starting with DISABLE_CHOWN : $DISABLE_CHOWN"

# root运行容器，容器里面一样root运行
if [ $USER_UID == '0' ]; then

    echo "-----root------Starting"

    USERNAMEROOT=root

    chown -R $USERNAMEROOT:$USERNAMEROOT /home/project
    chown -R $USERNAMEROOT:$USERNAMEROOT /home/opvscode

    #chmod +x /home/opvscode/server.sh
    ln -sf /home/$USERNAME/.nvm/versions/node/v$NODE_VERSION/bin/node /home/opvscode

    export HOME=/root

    echo "root:$USER_PASS" | chpasswd

    echo "-----------Starting sshd"
    /usr/sbin/sshd

    echo "-----------Starting smartide-python-v2-vscode"
    exec /home/smartide/.nvm/versions/node/v16.9.1/bin/node /home/opvscode/out/server-main.js --host 0.0.0.0 --without-connection-token

else

    #非root运行，通过传入环境变量创建自定义用户的uid,gid，否则默认uid,gid为1000
    echo "-----smartide------Starting"

    # 启动传UID=1000  不需要修改UID，GID值
    if [[ $USER_UID != 1000 ]]; then
        echo "-----smartide---usermod uid start---"$(date "+%Y-%m-%d %H:%M:%S")
        usermod -u $USER_UID $USERNAME
        find / -user 1000 -exec chown -h $USERNAME {} \;
        echo "-----smartide---usermod uid end---"$(date "+%Y-%m-%d %H:%M:%S")
    fi

    if [[ $USER_GID != 1000 ]]; then
        echo "-----smartide---usermod gid start---"$(date "+%Y-%m-%d %H:%M:%S")
        # groupmod -g $USER_GID $USERNAME
        groupmod -g $USER_GID --non-unique $USERNAME
        find / -group 1000 -exec chgrp -h $USERNAME {} \;
        echo "-----smartide---usermod gid end---"$(date "+%Y-%m-%d %H:%M:%S")
    fi

    export HOME=/home/$USERNAME
    # chmod g+rw /home
    chown -R $USERNAME:$USERNAME /home/project
    chown -R $USERNAME:$USERNAME /home/opvscode
    chown -R $USERNAME:$USERNAME /home/$USERNAME/.ssh

    #chmod +x /home/opvscode/server.sh

    echo "root:$USER_PASS" | chpasswd
    echo "smartide:$USER_PASS" | chpasswd

    # cp -r /root/.nvm /home/$USERNAME
    ln -sf /home/$USERNAME/.nvm/versions/node/v$NODE_VERSION/bin/node /home/opvscode

    echo "-----smartide------Starting sshd"
    # do not detach (-D), log to stderr (-e), passthrough other arguments
    # exec /usr/sbin/sshd -D -e "$@"
    /usr/sbin/sshd

    echo "-----smartide-----Starting gosu smartide-python-v2-vscode"
    exec su smartide -c "/home/smartide/.nvm/versions/node/v16.9.1/bin/node /home/opvscode/out/server-main.js --host 0.0.0.0 --without-connection-token"

fi
