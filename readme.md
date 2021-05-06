![overlay](resources/SC2-overlay.png)

Keep track of your basic game stats. Your win/loss record, win percentage, MMR gain/loss, and total win/loss for each gaming session.

This tool creates 6 text files in your current working directory:
* ZvP.txt
* ZvT.txt
* ZvZ.txt
* MMR-diff.txt
* winrate.txt
* totalWinLoss.txt

## useAPI

When a game finishes it saves your MMR to the replay file (the start of game MMR). [^1]

There's a setting called `useAPI`. It's optional, but it makes your MMR-diff more accurate by 1 game (~ 20 MMR). If set to false, I get your MMR data from the replay. This isn't good because you must play 2 games before it starts showing a ± MMR. Also, on your last game I don't know your live MMR as that is recorded at the end of the next game.

If set to true, then it's perfectly accurate. Your start MMR is set immediately, and as your last game finishes I also get your live MMR number.

## Usage

Place the .exe in any folder and run it. It will instantly close creating a template settings file (cfg.toml). Modify these 3 lines in `cfg.toml` and you're done. Rerun the program:
- name = 
- mainToon =
- dir = 

1. Find your starcraft2.com profile URL.
    - set name
    - set race [^2]
1. List all accounts you have for `name =`. [^3]
1. Set `mainToon =`. Use your profileID (last number in URL).
1. Set `dir =`. Use forward slash or double backslash for where your replays are saved.
1. Set `useAPI =` to true if you want the most accurate MMR calculator.

```sh
#     name - Put a comma-separated list of your SC2 accounts like in the example (url, name, race).
# mainToon - Choose only one profileID to use (only used if useAPI = true).
#      dir - Where to watch for new SC2 replays (use either a single slash, or a double backslash).

[account]
name = [ [ "https://starcraft2.com/en-gb/profile/1/1/1331332", "Gixxasaurus", "zerg" ] ]

# name = [ [ "https://starcraft2.com/en-gb/profile/1/1/1331332", "Gixxasaurus", "zerg" ],
#          [ "https://starcraft2.com/en-gb/profile/2/1/4545534", "Rairden", "zerg" ] ]

mainToon = "1331332"

[directory]
dir = "/home/erik/scratch/replays/"
# dir = "C:/Users/Erik/Downloads/reps/"
# dir = "C:\\Users\\Erik\\Downloads\\reps\\"

[settings]
updateTime = 1000
useAPI = false
OAuth2Creds = "http://108.61.119.116"
clientID = ""
clientSecret = ""
```

After you play SC2 or put a replay in your watch folder, the app will generate 6 .txt files you can use as overlays in something like OBS for streaming.

My OAuth2 credentials are used, but you can take 5 minutes to register your own Client ID and insert those values in the cfg.toml file.

[^1]: When you click "Score screen" data is sent back to blizzard. If both players sit in the replay (F10, w (rewind)) then the replay is not saved to disk therefore not reported to the blizzard API yet.
[^2]: I need your race here because your account will have a distinct `ladderID` for each 1v1 race.
[^3]: If you only have one account URL, then use the first name= template. Lines starting w/ hashtag (#) are ignored.
