#!/bin/bash

USER_PASS=${LOCAL_USER_PASSWORD:-"smartide123.@IDE"}
echo "Starting with USER_PASS : $USER_PASS"
echo "root:$USER_PASS" | chpasswd

exec /usr/sbin/sshd -D -e "$@"
