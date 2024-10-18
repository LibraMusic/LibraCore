# Libra

Libra is a new, open, and extensible music service. Libra does what you want, how you want.

## Setup Steps

Before anything else, you need to create the database. Install PostgreSQL and run the following commands:

```bash
sudo -u postgres createuser -P libra
sudo -u postgres createdb -O libra -E UTF-8 libra
```

Dependencies:

- PostgreSQL
- yt-dlp
- ytmusicapi
- FFmpeg

## Roadmap

In the future, the project aims for the following:

- [ ] A simple but powerful backend & API
- [ ] A beautiful and accessible frontend for all platforms, both web and native
- [ ] An opt-in playback syncing mechanism to share what is currently playing between applications on either one or every device
- [ ] A plugins system, so you can extend Libra even more than normal
- [ ] An Amazon Alexa skill
- [ ] An [Obsidian](https://obsidian.md/) plugin to control your music while taking notes
