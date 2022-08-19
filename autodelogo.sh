#!/bin/bash

echo ------------------spliting------------------------
#00:00:00
ffmpeg -i download.mp4 -vcodec copy -acodec copy -t $TIME tmp1.mp4
ffmpeg -i download.mp4 -vcodec copy -acodec copy -ss $TIME tmp2.mp4

rm -rf download.mp4

echo ------------------delogoing------------------------
# "delogo=x=432:y=44:w=1060:h=108"
ffmpeg -i tmp1.mp4 -vf $1 -c:a copy tmp3.mp4
