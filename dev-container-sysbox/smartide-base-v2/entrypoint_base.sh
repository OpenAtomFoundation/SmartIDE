#!/bin/bash

USER_UID=${LOCAL_USER_UID:-1000}
USER_GID=${LOCAL_USER_GID:-1000}
USER_PASS=${LOCAL_USER_PASSWORD:-"123qwe"}
USERNAME=smartide
echo "Starting with USER_UID : $USER_UID"
echo "Starting with USER_GID : $USER_GID"
echo "Starting with USER_PASS : $USER_PASS"

# root运行容器，容器里面一样root运行
if [ $USER_UID == '0' ]; then

    echo "-----root------Starting"

    USERNAMEROOT=root

    chown -R $USERNAMEROOT:$USERNAMEROOT /home/project
    chown -R $USERNAMEROOT:$USERNAMEROOT /home/opvscode

    export HOME=/root

    echo "root:$USER_PASS" | chpasswd
    echo "-----------Starting sshd"
    # 后面不加$@容器会自动退出
    exec /usr/sbin/sshd -D -e "$@"

else

    #非root运行，通过传入环境变量创建自定义用户的uid,gid
    echo "-----smartide------Starting"

    export HOME=/home/$USERNAME

    chown -R $USERNAME:$USERNAME /home/project
    chown -R $USERNAME:$USERNAME /home/opvscode

    echo "root:$USER_PASS" | chpasswd
    echo "smartide:$USER_PASS" | chpasswd
    echo "-----smartide------Starting sshd"
    exec /usr/sbin/sshd -D -e "$@"

fi
