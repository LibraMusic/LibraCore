import json
import os
import sys

from ytmusicapi import YTMusic
import yt_dlp

if len(sys.argv) < 2:
    print("Error: Missing action argument")
    sys.exit()

script_path = os.path.dirname(os.path.realpath(__file__))

ytmusic = YTMusic()

action = sys.argv[1]

args = {}
for arg in sys.argv[2:]:
    if "=" in arg:
        key, value = arg.split("=", 1)
        args[key] = value

if action == "search":
    if "query" not in args:
        print("Error: Missing query argument")
        sys.exit()
    query = args["query"]
    limit_str = args.get("limit", "20")
    if limit_str == "all":
        limit = None
    else:
        limit = int(limit_str)
    filters = json.loads(args.get("filters", "{}"))
    allow_videos = filters.get("allow_videos", False)

    searched_types = filters.get("types", ["tracks", "albums", "artists"])
    if "tracks" in searched_types and allow_videos and "videos" not in searched_types:
        searched_types.append("videos")

    search_result_lists = []
    for searched_type in searched_types:
        search_result_lists.append(
            ytmusic.search(
                query,
                filter=searched_type if searched_type != "tracks" else "songs",
                limit=limit,
                ignore_spelling=True,
            )
        )

    # Merge the results evenly (e.g. 1st result from each type, then 2nd result from each type, etc.)
    search_results = []
    while len(search_results) < limit:
        for search_result_list in search_result_lists:
            if len(search_result_list) == 0:
                continue
            search_results.append(search_result_list.pop(0))
            if len(search_results) >= limit:
                break

    print(json.dumps(search_results))
elif action == "lyrics":
    if "id" not in args:
        print("Error: Missing id argument")
        sys.exit()
    if args["id"].startswith("MPLY"):
        lyrics_id = args["id"]
    else:
        video_id = args["id"]
        lyrics_id = ytmusic.get_watch_playlist(video_id)["lyrics"]
    lyrics = ytmusic.get_lyrics(lyrics_id)

    # YouTube doesn't specify the language of the lyrics, so we'll just say it's unknown.
    print(json.dumps({"unknown": "txt\n" + lyrics["lyrics"].replace("\r\n", "\n")}))
elif action == "subtitles":
    if "id" not in args:
        print("Error: Missing id argument")
        sys.exit()
    video_id = args["id"]
    url = f"https://www.youtube.com/watch?v={video_id}"

    video_directory = os.path.join(script_path, video_id)
    file_path = os.path.join(video_directory, "subtitles")
    os.makedirs(file_path, exist_ok=True)

    ydl_opts = {
        "outtmpl_na_placeholder": "",
        "outtmpl": {"subtitle": os.path.join(file_path, "%(ext)s")},
        "postprocessors": [
            {
                "key": "FFmpegSubtitlesConvertor",
                "format": "vtt",
                "when": "before_dl",
            }
        ],
        "skip_download": True,
        "subtitleslangs": ["all", "-live_chat"],
        "writesubtitles": True,
        "quiet": True,
        "noprogress": True,
    }

    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        ydl.download([url])

    subtitles = {}
    for file in os.listdir(file_path):
        with open(os.path.join(file_path, file), "r") as f:
            subtitles[file.split(".")[-2]] = file.split(".")[-1] + "\n" + f.read()
        os.remove(os.path.join(file_path, file))

    if not os.listdir(file_path):
        os.rmdir(file_path)
    if not os.listdir(video_directory):
        os.rmdir(video_directory)

    print(json.dumps(subtitles))
elif action == "content":
    if "id" not in args:
        print("Error: Missing id argument")
        sys.exit()
    video_id = args["id"]

    content_type = args.get("type", "audio")

    if content_type == "audio":
        ydl_opts = {
            "format": "bestaudio/best",
            "multiple_audiostreams": True,
            "outtmpl": "-",
            "quiet": True,
            "noprogress": True,
        }
    else:
        ydl_opts = {
            "format": "bestvideo+bestaudio/best",
            "multiple_audiostreams": True,
            "multiple_videostreams": True,
            "outtmpl": "-",
            "quiet": True,
            "noprogress": True,
        }

    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        ydl.download([f"https://www.youtube.com/watch?v={video_id}"])
elif action in ("track", "video"):
    if "id" not in args:
        print("Error: Missing id argument")
        sys.exit()
    video_id = args["id"]
    video = ytmusic.get_song(video_id)

    watch_playlist = ytmusic.get_watch_playlist(video_id)
    track = watch_playlist["tracks"][0]
    track["lyricsId"] = watch_playlist["lyrics"]

    print(json.dumps({"video": video, "track": track}))
elif action == "album":
    if "id" not in args:
        print("Error: Missing id argument")
        sys.exit()
    album_id = args["id"]
    album = ytmusic.get_album(album_id)

    print(json.dumps(album))
elif action == "artist":
    if "id" not in args:
        print("Error: Missing id argument")
        sys.exit()
    artist_id = args["id"]
    artist = ytmusic.get_artist(artist_id)

    print(json.dumps(artist))
elif action == "playlist":
    if "id" not in args:
        print("Error: Missing id argument")
        sys.exit()
    playlist_id = args["id"]
    playlist = ytmusic.get_playlist(playlist_id)

    print(json.dumps(playlist))
