![overlay](resources/SC2-overlay.png)

Keep track of your basic game stats. Your win/loss record, win percentage, MMR gain/loss, and total win/loss for each gaming session.

This tool creates 6 text files in your current working directory:
* ZvP.txt
* ZvT.txt
* ZvZ.txt
* MMR-diff.txt
* winrate.txt
* totalWinLoss.txt

It only works for zerg players, or 3 of the 6 matchups.

## useAPI

When a game finishes it saves your MMR to a .SC2Replay file (the start of game MMR). [^1]

There's a setting called `useAPI`. It's optional, but it makes your `MMR-diff` more accurate by 1 game (~ 20 MMR). If set to false, I get your MMR data from the replay. This isn't good because you must play 2 games before it starts showing a ± MMR. Also, on your last game I don't know your live MMR. [^2]

If set to true, then it's perfectly accurate. Your start MMR is set immediately, and as your last game finishes I also get your live MMR number.

## Usage

```
# Lines starting with a hashtag (#) are ignored.
# The minimum configuration is 2 fields (name=, dir=)            for useAPI = false.
# The minimum configuration is 3 fields (name=, dir=, mainToon=) for useAPI = true.
```

Place the .exe in any folder and run it. It will instantly close creating a template settings file (cfg.toml). Modify these 3 lines in `cfg.toml` and you're done. Rerun the program:
- name = 
- mainToon =
- dir = 

1. Find your starcraft2.com profile URL.
    - set name
    - set race [^3]
1. List all accounts you have for `name =`. [^4]
1. Set `mainToon =`. Use your profileID (last number in URL).
1. Set `dir =`. Use forward slash or double backslash for where your replays are saved.
1. Set `useAPI =` to true if you want the most accurate MMR calculator.

You won't use this, but here is a minimum configuration cfg.toml file. Change 2 lines (name=, dir=).

```sh
[account]
name = [ [ "https://starcraft2.com/en-gb/profile/1/1/1331332", "Gixxasaurus", "zerg" ] ]
mainToon = ""

[directory]
dir = "C:/Users/Erik/Downloads/reps"

[settings]
updateTime = 1000
useAPI = false
OAuth2Creds = ""
clientID = ""
clientSecret = ""
```

After you play SC2 or put a replay in your watch folder, the app will generate 6 .txt files you can use as overlays in something like OBS for streaming.

My OAuth2 credentials are used, but you could take 5 minutes to register your own Client ID [^5] and insert those 2 values at the very bottom of the cfg.toml file.

## How it works

The program only keeps track 1v1 ranked ladder. Replays are ignored if they are computer A.I. or 1v1 custom. Unranked replays behave identical to ranked replays (updates stats, but MMR should not change).

- all files not ending with .SC2Replay are ignored
- subfolders are ignored
- computer A.I. replays are ignored
- custom 1v1 replays are ignored
- deleting 1+ files doesn't recalculate your stats

Say you play your first game and it ends. It shows you're +15 MMR now. You want to practice against the A.I. and delete the replay afterwards. Once the A.I. game is over, the program ignores the new replay and doesn't change your stats. Also, you can delete the A.I. replay and it doesn't recalculate any stats. It ignores deleting *any* type of file. If you delete all .SC2Replay files then it resets all stats to 0.

## The code

```sh
wc **/*.go | sort -k 1n
   31    84  1106 api/laddersummary/ladder-summary.go
   52   143   994 sc2replay_test.go
   58   142  1820 api/ladder/ladder.go
  115   346  2197 stats.go
  161   537  4598 api.go
  185   631  4859 filemgr.go
  429  1115  8909 sc2replay.go
 1031  2998 24483 total
 ```

Make a plantuml from source code

```sh
goplantuml -recursive -show-aggregations -show-compositions -aggregate-private-members -show-implementations -show-connection-labels $GOPATH/src/sc2-replay-go > file.puml
```

creates this UML [diagram](/resources/UML-sc2rep.png).

[^1]: When you click "Score screen" or "Rewind" data is sent back to blizzard.  
[^2]: The start of game MMR, not the end result MMR.  
[^3]: I need your race here because your account will have a distinct `ladderID` for each 1v1 race.  
[^4]: If you only have one account URL, then use the first name= template.  
[^5]: <https://develop.battle.net/documentation/guides/getting-started> If I die, or my web server loses power, then to continue using this app you need to get your own free clientID/clientSecret.  
