#!/bin/bash

# Usage: build_container_image.sh [-h] [-r]
#        script to build image from helm-cli
#
# Options:
#   -h  Show this help text
#   -r  Set repotag (default: zanattabruno/ric-plt-vespamgr:energy-saver)
#
# Description:
#   This script builds a Docker image from the helm-cli. It allows you to specify a custom repotag using the -r option, 
#   otherwise it uses the default repotag. The built image is then pushed to the specified repotag.
#
# Example usage:
#   $ build_container_image.sh -r myrepo/myimage:latest
#   $ build_container_image.sh -h

usage="$(basename "$0") [-h] [-r] -- script to build image from helm-cli

where:
    -h  show this help text
    -r  set repotag | default zanattabruno/ric-plt-vespamgr:energy-saver"

repotag=zanattabruno/ric-plt-vespamgr:energy-saver

if [ "$*" == "" ]; then
    echo "No flag is passed using default repotag $repotag"
fi

while getopts :hr: flag
do
    case "${flag}" in
        r) repotag=${OPTARG};;
        h) echo "$usage"
        exit 0;;
    esac
done

docker build . -t $repotag && docker push $repotag
