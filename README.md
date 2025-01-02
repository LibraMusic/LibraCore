# Libra

Libra is a new, open, and extensible music service. Libra does what you want, how you want.

The current goal is to have LibraCore at a pre-releasable state by the end of March, 2025.

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

-   `SQLite` or `PostgreSQL`
-   `yt-dlp` and `ytmusicapi` for the YouTube source
-   `FFmpeg`

## Development

To run all unit tests, run `make test`.
To run the integration tests, run `make test_integration`
