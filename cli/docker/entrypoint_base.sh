#!/bin/bash
# SPDX-License-Identifier: GPL-3.0-or-later
# Copyright (C) 2023 SmartIDE Server & leansoftx.com

exec /usr/sbin/sshd -D -e "$@"
