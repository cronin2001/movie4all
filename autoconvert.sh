#!/bin/bash

echo ------------------forcetbn-------------------------
ffmpeg -i tmp3.mp4 -strict -2 -video_track_timescale $1 tmp4.mp4

echo ------------------merging------------------------
echo file tmp4.mp4 > mylist.txt && echo file tmp2.mp4 >> mylist.txt
ffmpeg -f concat -i mylist.txt -c copy output.mp4

rm -rf tmp* mylist.txt

mkdir main

echo ------------------converting------------------------
ffmpeg -i output.mp4 -codec: copy -start_number 0 -hls_time 10 -hls_list_size 0 -f hls ./main/main.m3u8

rm -rf output.mp4

# mv main upload

# cd /content/drive/MyDrive/upload && cp main.go /content/upload

# echo ------------------uploading------------------------
# cd /content/upload && go mod init upload && go mod tidy && go run main.go main

# cd /content && rm -rf upload
