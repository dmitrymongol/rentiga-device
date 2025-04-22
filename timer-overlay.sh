#!/bin/bash

# Параметры устройства
DEVICE="/dev/video0"
RESOLUTION="1920x1080"  # Подберите под ваше устройство

# Запуск FFmpeg с фильтром таймера
ffmpeg -f v4l2 -input_format mjpeg -framerate 30 -video_size $RESOLUTION \
-i $DEVICE \
-vf "drawtext=fontfile=/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf: \
text='%{localtime\\:%T}': fontcolor=white@0.9: fontsize=100: \
box=1: boxcolor=black@0.5: boxborderw=10: \
x=(w-text_w)/2: y=(h-text_h)*0.8" \
-vcodec rawvideo -pix_fmt bgr0 -f sdl "Video Preview"