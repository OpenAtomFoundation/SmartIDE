



# "https://download.jetbrains.com/idea/ideaIC-2021.2.3.tar.gz"
# containerName=${1:-projector-idea-c}
# downloadUrl=${2:-https://download.jetbrains.com/idea/ideaIC-2019.3.5.tar.gz}
# # build container:
# DOCKER_BUILDKIT=1 docker build --progress=plain -t "$containerName" --build-arg buildGradle=false --build-arg "downloadUrl=$downloadUrl" -f Dockerfile ..

# docker build --progress=plain -t $(ImageName):$(ImageTag) --build-arg "downloadUrl=$(IdeDownloadUrl)" -f Dockerfile ..

echo "----------------执行git clone start\n"

git clone https://github.com/JetBrains/projector-server.git

echo "----------------执行git clone end\n"

