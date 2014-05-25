#!/bin/bash

if [[ -z "$1" ]]
  then
  echo "specify a video filename"
  exit 1
fi

command="ffmpeg -i $1 -y -f mp3 -ab 192000 -vn $1.mp3"
echo $command
