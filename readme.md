![overlay](resources/SC2-overlay.png)

## Usage

Place the .exe in any folder. Double-click `sc2-replay-go.exe` and it will instantly close creating a template settings file (cfg.toml). Modify two lines in `cfg.toml` and you're done.

Here is an example `cfg.toml` file if you have only one SC2 name.

If your SC2 name is `Reynor` and you store your replays in the folder `E:\SC2\replayBackup`, then this config file would work:

```sh
# name - Put a comma-separated list of your SC2 player names.
[account]
name = [ "Reynor" ]

# dir - Where to watch for new SC2 replays (use either a single slash, or a double backslash).
[directory]
dir = "E:\\SC2\\replayBackup\\"
```

You can use forward slash or double backslash for the path.  
Here is another example cfg.toml file with my two accounts:

```sh
[account]
name = [ "Gixxasaurus", "Rairden" ]

[directory]
dir = "C:/Users/Erik/Documents/StarCraft II/Accounts/63960513/1-S2-1-1331332/Replays/Multiplayer/"
```

After you play SC2 or put a replay in your watch folder, the app will generate 5 .txt files you can use as overlays in something like OBS for streaming:

* ZvP.txt
* ZvT.txt
* ZvZ.txt
* MMR-diff.txt
* winrate.txt

