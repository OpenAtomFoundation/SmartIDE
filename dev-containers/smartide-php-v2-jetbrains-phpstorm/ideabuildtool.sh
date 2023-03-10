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



# "https://download.jetbrains.com/idea/ideaIC-2021.2.3.tar.gz"
# containerName=${1:-projector-idea-c}
# downloadUrl=${2:-https://download.jetbrains.com/idea/ideaIC-2019.3.5.tar.gz}
# # build container:
# DOCKER_BUILDKIT=1 docker build --progress=plain -t "$containerName" --build-arg buildGradle=false --build-arg "downloadUrl=$downloadUrl" -f Dockerfile ..

# docker build --progress=plain -t $(ImageName):$(ImageTag) --build-arg "downloadUrl=$(IdeDownloadUrl)" -f Dockerfile ..

echo "----------------执行git clone start\n"

git clone https://github.com/JetBrains/projector-server.git

echo "----------------执行git clone end\n"

