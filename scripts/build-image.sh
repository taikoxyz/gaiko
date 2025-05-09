#!/usr/bin/env bash
set -eo pipefail

# docker build
read -p "Do you want to build local(0) or xxx(1): " docker_type
case "$docker_type" in
0 | local)
    image_name=gaiko-local
    target_dockerfile=docker/Dockerfile.local
    ;;
1)
    echo "not supported yet"
    exit 1
    ;;
*)
    echo "unknown gaiko type to build, use 'local' as default"
    image_name=gaiko-local
    target_dockerfile=docker/Dockerfile.local
    ;;
esac

read -p "Image version: " tag
if [ -z "$tag" ]; then
    echo "Image version is missing, use 'latest' as default"
    tag=latest
fi

docker buildx build . \
    -f $target_dockerfile \
    --load \
    --platform linux/amd64 \
    -t $image_name:$tag \
    --build-arg TARGETPLATFORM=linux/amd64 \
    --progress=plain
