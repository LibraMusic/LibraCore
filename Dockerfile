FROM python:3.13-alpine

RUN apk add --no-cache ffmpeg
RUN pip install --no-cache-dir yt-dlp ytmusicapi

WORKDIR /app

EXPOSE 8080

COPY libra ./
ENTRYPOINT ["./libra", "server"]
