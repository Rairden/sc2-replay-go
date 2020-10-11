The first game (Zed) I played I had no MMR field in the json.  
The second game (silver) my MMR was set to -36400 MMR.  
Two to three of my first placement matches had that -36400 MMR.

| matchup winner | player      | mmr    |
| :------------- | :---------- | -----: |
| ZvT senya      | senya       | 2947   |
|                | Rairden     | 3046   |
| ZvP Rairden    | Alby        | 2971   |
|                | Rairden     | 3039   |
| ZvP PiotrekS   | PiotrekS    | 3153   |
|                | Rairden     | 3045   |
| ZvP silver     | silver      | 1684   |
|                | Gixxasaurus | -36400 |
| ZvZ Zed        | Zed         | 2594   |
|                | Gixxasaurus |        |
| ZvZ Rairden    | Rairden     | 3059   |
|                | Brubbster   | 3061   |
| ZvT vapit      | Rairden     | 3049   |
|                | vapit       | 3068   |
| ZvT tobe       | tobe        | 3315   |
|                | Rairden     | 3059   |

The correct output for my record here is:

ZvP 1-2  
ZvT 0-3  
ZvZ 1-1

My SC2 player names and ID's are:

* Gixxasaurus, 1331332
* Rairden, 4545534

## Go vs Java performance

There's 2-3 tools to decode blizzard MPQ replay files. I previously did this in Java, and used a python library called [sc2reader](https://pypi.org/project/sc2reader/).

Go is much faster.

Here is Blizzard's offical low-level tool written in python linking the Java and Go implmentations at the very bottom of [github](https://github.com/Blizzard/s2protocol#ports-and-related-projects).

I benchmarked sc2reader against icza's `s2prot` on 100 replays.

| decoder   | lang   | runtime (ms) |
| --------- | ------ | -----------: |
| sc2reader | Python | 580          |
| s2prot    | Go     | 167          |

I used `load_level=2` in sc2reader (2 of 5). I did the same with Go, and only grabbed the minimal decoding information (145 lines of json).

| flag             | lines   | size |
| ---------------: | ------: | ---- |
| header.json      | 69      | 1.6K |
| details.json     | 145     | 4.2K |
| msgevts.json     | 291     | 5.6K |
| attrevts.json    | 890     | 21K  |
| initdata.json    | 2,081   | 54K  |
| trackerevts.json | 77,105  | 2.1M |
| gameevts.json    | 307,376 | 6.0M |


What that looked like in python:

```python
replay = sc2reader.load_replay(filepath, load_level=2)  # minimal decode
```

And Go:

```go
r, err := rep.NewFromFile("rep.SC2Replay")                           // full decode
r, err := rep.NewFromFileEvts("rep.SC2Replay", false, false, false)  // minimal
```

