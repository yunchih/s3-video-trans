FROM jrottenberg/ffmpeg:3.4-alpine

MAINTAINER dlc-hackers
USER root

ADD static-minio-video-transcoder /transcoder 
ENTRYPOINT ["/transcoder"]
