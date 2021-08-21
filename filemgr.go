package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	_, r, _, _      = runtime.Caller(0) // linux (testing)
	s, _            = os.Executable()   // win10
	currDir         = path.Dir(s) + filepath.Join("/")
	cfgToml         = currDir + "cfg.toml"
	xvpTxt          = currDir + "xvp.txt"
	xvtTxt          = currDir + "xvt.txt"
	xvzTxt          = currDir + "xvz.txt"
	mmrDiffTxt      = currDir + "MMR-diff.txt"
	winrateTxt      = currDir + "winrate.txt"
	totalWinLossTxt = currDir + "totalWinLoss.txt"
	logDir          = currDir + "log" + filepath.Join("/")
	logFile         = logDir + "errors.log"
	cfg             settings
)

// the cfg.toml file
type settings struct {
	mainToon, dir string
	updateTime    int64
	useAPI        bool
	OAuth2Creds   string
	clientID      string
	clientSecret  string
}

// absolutePath is your cfg.toml file
func setup(absolutePath string) *player {
	os.Mkdir(logDir, os.ModePerm)
	logs, _ := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	log.SetOutput(logs)

	player := &player{
		[2]uint8{}, [2]uint8{}, [2]uint8{}, [2]uint8{},
		0, 0,
		make(map[string]*profile),
	}

	if cfgExists() {
		config, _ := toml.Load(cfgToString(absolutePath))
		toons := config.Get("account.name").([]interface{})
		mainToon := config.Get("account.mainToon").(string)
		dir := config.Get("directory.dir").(string)
		useAPI := config.Get("settings.useAPI").(bool)
		updateTime := config.Get("settings.updateTime").(int64)
		OAuth2Creds := config.Get("settings.OAuth2Creds").(string)

		clientID := config.Get("settings.clientID").(string)
		clientSecret := config.Get("settings.clientSecret").(string)

		dir = filepath.Clean(dir)
		dir += filepath.Join("/")
		updateTime = setUpdateTime(updateTime)

		for i := range toons {
			arr := toons[i].([]interface{})

			url := arr[0].(string)
			name := arr[1].(string)
			race := arr[2].(string)
			race = strings.ToLower(race)

			split := strings.Split(url, "/")

			regionID := split[5]
			realmID := split[6]
			profileID := split[7]

			region := getRegion(regionID, realmID)

			profile := &profile{
				url, name, race,
				regionID, realmID, profileID, "",
				region,
			}

			player.profile[profileID] = profile
		}

		cfg = settings{
			mainToon, dir,
			updateTime,
			useAPI,
			OAuth2Creds,
			clientID,
			clientSecret,
		}
	} else {
		writeData(cfgToml, myToml)
		fmt.Println("Now setup your cfg.toml file.")
		os.Exit(0)
	}

	return player
}

// https://develop.battle.net/documentation/guides/regionality-and-apis
func getRegion(reg, realmID string) string {
	switch reg {
	case "1":
		return "us"
	case "2":
		return "eu"
	case "3":
		if realmID == "1" {
			return "kr"
		}
		if realmID == "2" {
			return "tw"
		}
	}
	return ""
}

func cfgExists() bool {
	_, err := os.Open(cfgToml)
	return err == nil
}

func cfgToString(absolutePath string) string {
	b, err := ioutil.ReadFile(absolutePath)
	if err != nil {
		fmt.Printf("File not found: '%v'", b)
	}
	return string(b)
}

// Blizzard: "API clients are limited to 36,000 requests per hour at a rate of 100 requests per second."
// limit file watcher update from 0.1 sec to 10 seconds
func setUpdateTime(t int64) int64 {
	if t < 100 {
		return 100
	}
	if t > 10000 {
		return 10000
	}
	return t
}

var myToml = `# Lines starting with a hashtag (#) are ignored.
# The minimum configuration is 2 fields (name=, dir=)            for useAPI = false.
# The minimum configuration is 3 fields (name=, dir=, mainToon=) for useAPI = true.
#
#     name - Put a comma-separated list of your SC2 accounts like in the example (url, name, race).
# mainToon - Choose only one profileID to use (only used if useAPI = true).
#      dir - Where to watch for new SC2 replays (can use slash or backslash).

[account]
name = [ [ 'https://starcraft2.com/en-gb/profile/1/1/1331332', 'Gixxasaurus', 'zerg' ] ]

# name = [ [ 'https://starcraft2.com/en-gb/profile/1/1/1331332', 'Gixxasaurus', 'zerg' ],
#          [ 'https://starcraft2.com/en-gb/profile/2/1/4545534', 'Rairden', 'zerg' ] ]

mainToon = '1331332'

[directory]
dir = '/home/erik/scratch/replays'
# dir = 'E:\SC2\replayBackup'
# dir = 'E:/SC2/replayBackup'

#  updateTime - How often to check dir if a new .SC2Replay file is created. Time in milliseconds. Range of 100 to 10000.
#      useAPI - Whether or not to get your MMR from the battlenet API (default: true).
# OAuth2Creds - Do not change. Where to get OAuth2 credentials in order to use battlenet API.
#    clientID - Optional. Fill in ID/pass if you registered your own Client (best option).
[settings]
updateTime = 1000
useAPI = true
OAuth2Creds = 'http://108.61.119.116'
clientID = ''
clientSecret = ''
`
