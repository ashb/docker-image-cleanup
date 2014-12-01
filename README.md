# Docker-image-cleanup

Docker image cleanup tool.

Removes old and extraneous images from localhost

## Install

go install github.com/lever/docker-image-cleanup

## Usage

docker-image-cleanup -h # Show help
docker-image-cleanup -a 30 -k 10 # Delete images over 30 days old, keep at least 10 per image name

## Config

By default this tool deletes images which are both: more than 14 days old and greater than the fifth-most-recently downloaded version of each unique image name.

Use the -a flag to change how many days old an image must be before being considered for deletion.
Use the -k flag to configure the minimum number of versions to keep per image name.
