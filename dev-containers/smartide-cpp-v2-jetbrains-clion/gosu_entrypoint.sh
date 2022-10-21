#!/bin/bash
###
 # @Author: kenan
 # @Date: 2022-05-24 14:37:27
 # @LastEditors: kenan
 # @LastEditTime: 2022-05-24 15:11:47
 # @FilePath: /smartide/dev-containers/smartide-python-v2-jetbrains-clion/gosu_entrypoint.sh
 # @Description: 
 # 
 # Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved. 
### 

USER_UID=${LOCAL_USER_UID:-1000}
USER_GID=${LOCAL_USER_GID:-1000}
USER_PASS=${LOCAL_USER_PASSWORD:-"smartide123.@IDE"}
USERNAME=smartide

echo "gosu_entrypoint.sh"
echo "Starting with USER_UID : $USER_UID"
echo "Starting with USER_GID : $USER_GID"
echo "Starting with USER_PASS : $USER_PASS"

# root运行容器，容器里面一样root运行
if [ $USER_UID == '0' ]; then

    echo "-----root------Starting"

    USERNAMEROOT=root

    chown -R $USERNAMEROOT:$USERNAMEROOT /home/project
    #chown -R $USERNAMEROOT:$USERNAMEROOT /home/opvscode

    #chmod +x /home/opvscode/server.sh
    #ln -sf /home/$USERNAME/.nvm/versions/node/v$NODE_VERSION/bin/node /home/opvscode

    export HOME=/root

    echo "root:$USER_PASS" | chpasswd

    echo "-----------Starting sshd"
    /usr/sbin/sshd

    echo "-----------Starting ide CLion"
    # exec run.sh "$@"
    cd /home/smartide && exec ./run.sh

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
    chown -R $USERNAME:$USERNAME /home/$USERNAME/.ssh

    #chown -R $USERNAME:$USERNAME /home/opvscode
    #chmod +x /home/opvscode/server.sh

    echo "root:$USER_PASS" | chpasswd
    echo "smartide:$USER_PASS" | chpasswd

    # cp -r /root/.nvm /home/$USERNAME
    #ln -sf /home/$USERNAME/.nvm/versions/node/v$NODE_VERSION/bin/node /home/opvscode

    echo "-----smartide------Starting sshd"
    # do not detach (-D), log to stderr (-e), passthrough other arguments
    #exec /usr/sbin/sshd -D -e "$@"
    /usr/sbin/sshd
    
    echo "-----smartide-----Starting gosu CLion"
    cd /home/smartide && exec gosu smartide ./run.sh

fi
