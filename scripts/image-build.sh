#!/bin/bash

# build the docker image using the binaries, if you want to reduce the size of the image,
# use upx to compress the binaries before building the image.

serverName="user_server_go"
# image name of the service, prohibit uppercase letters in names.
IMAGE_NAME="user-server-go/user-server-go"
# Dockerfile file directory
DOCKERFILE_PATH="scripts/build"
DOCKERFILE="${DOCKERFILE_PATH}/Dockerfile"

# image repo address, REPO_HOST="ip or domain", passed in via the first parameter
REPO_HOST=$1
if [ "X${REPO_HOST}" = "X" ];then
        echo "param 'repo host' cannot be empty, example: ./image-build.sh hub.docker.com v1.0.0"
        exit 1
fi
# the version tag, which defaults to latest if empty, is passed in via the second parameter
TAG=$2
if [ "X${TAG}" = "X" ];then
        TAG="latest"
fi
# image name and tag
IMAGE_NAME_TAG="${REPO_HOST}/${IMAGE_NAME}:${TAG}"

# binary executable files
BIN_FILE="cmd/${serverName}/${serverName}"
# configuration file directory
CONFIG_PATH="configs"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BIN_FILE} cmd/${serverName}/*.go
mv -f ${BIN_FILE} ${DOCKERFILE_PATH}
mkdir -p ${DOCKERFILE_PATH}/${CONFIG_PATH} && cp -f ${CONFIG_PATH}/${serverName}.yml ${DOCKERFILE_PATH}/${CONFIG_PATH}

# compressing binary file
#cd ${DOCKERFILE_PATH}
#upx -9 ${serverName}
#cd -

echo "docker build -f ${DOCKERFILE} -t ${IMAGE_NAME_TAG} ${DOCKERFILE_PATH}"
docker build -f ${DOCKERFILE} -t ${IMAGE_NAME_TAG} ${DOCKERFILE_PATH}


if [ -f "${DOCKERFILE_PATH}/${serverName}" ]; then
    rm -f ${DOCKERFILE_PATH}/${serverName}
fi

if [ -d "${DOCKERFILE_PATH}/configs" ]; then
    rm -rf ${DOCKERFILE_PATH}/configs
fi

# delete none image
noneImages=$(docker images | grep "<none>" | awk '{print $3}')
if [ "X${noneImages}" != "X" ]; then
  docker rmi ${noneImages} > /dev/null
fi
exit 0
