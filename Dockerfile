FROM python:3.14-alpine

ARG TARGETPLATFORM

RUN apk add --no-cache ffmpeg
RUN pip install --no-cache-dir yt-dlp ytmusicapi

EXPOSE 8080

COPY $TARGETPLATFORM/libra /libra
ENTRYPOINT ["/libra"]
CMD ["server"]
