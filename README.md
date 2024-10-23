# Libra

Libra is a new, open, and extensible music service. Libra does what you want, how you want.

## Setup Steps

Before anything else, you need to set up the database. Install one of the following databases and follow its steps:

<details>
<summary>SQLite</summary>

No additional steps are needed to use SQLite.
</details>

<details>

<summary>PostgreSQL</summary>

To create the PostgreSQL database, run the following commands:

```bash
sudo -u postgres createuser -P libra
sudo -u postgres createdb -O libra -E UTF-8 libra
```
</details>

Dependencies:

- `SQLite` or `PostgreSQL`
- `yt-dlp` and `ytmusicapi` for the YouTube source
- `FFmpeg`

## Development

To run all tests, run `make test`.
If you don't have every supported database type installed, you can skip database tests by setting the following environment variables:
- `SKIP_SQLITE_TESTS`
- `SKIP_POSTGRES_TESTS`

## Roadmap

In the future, the project aims for the following:

- [ ] A simple but powerful backend & API
- [ ] A beautiful and accessible frontend for all platforms, both web and native
- [ ] An opt-in playback syncing mechanism to share what is currently playing between applications on either one or every device
- [ ] A plugins system, so you can extend Libra even more than normal
- [ ] An Amazon Alexa skill
- [ ] An [Obsidian](https://obsidian.md/) plugin to control your music while taking notes
