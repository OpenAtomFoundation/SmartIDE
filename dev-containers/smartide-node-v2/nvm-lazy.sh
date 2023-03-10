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

export NVM_DIR="$HOME/.nvm"

node_versions=("$NVM_DIR"/versions/node/*)

if (( "${#node_versions[@]}" > 0 )); then
    PATH="$PATH:${node_versions[$((${#node_versions[@]} - 1))]}/bin"
fi

if [ -s "$NVM_DIR/nvm.sh" ]; then
    # load the real nvm on first use
    nvm() {
        # shellcheck disable=SC1090,SC1091
        source "$NVM_DIR"/nvm.sh
        nvm "$@"
    }
fi

if [ -s "$NVM_DIR/bash_completion" ]; then
    # shellcheck disable=SC1090,SC1091
    source "$NVM_DIR"/bash_completion
fi
