# Docker-image-cleanup

Docker image cleanup tool.

Removes old and extraneous images from localhost

## Install

go install github.com/lever/docker-image/cleanup

## Usage

docker-image-cleanup

## Config

By default this tool deletes images which are both: more than 14 years old and greater than the fifth-most-recently downloaded version of a particular image.

Use the -a flag to change how many days old a image will be kept for
Use the -k flag to configure the minimum number of versions per image name to keep.
