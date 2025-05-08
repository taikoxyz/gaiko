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
    echo "unknown proof type to build"
    exit 1
    ;;
esac

read -p "Image version: " tag
if [ -z "$tag" ]; then
    echo "Image version is required"
    exit 1
fi

docker buildx build . \
    -f $target_dockerfile \
    --load \
    --platform linux/amd64 \
    -t $image_name:$tag \
    --build-arg TARGETPLATFORM=linux/amd64 \
    --progress=plain
