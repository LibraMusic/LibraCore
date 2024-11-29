FROM python:3.12

RUN pip install --no-cache-dir yt-dlp ytmusicapi

WORKDIR /app

EXPOSE 8080

COPY libra ./
ENTRYPOINT ["./libra", "server"]
