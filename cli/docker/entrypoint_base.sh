#!/bin/bash

exec /usr/sbin/sshd -D -e "$@"
