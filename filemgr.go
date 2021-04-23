package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	_, r, _, _ = runtime.Caller(0) // linux (testing)
	s, _       = os.Executable()   // win10
	currDir    = path.Dir(s) + filepath.Join("/")
	cfgToml    = currDir + "cfg.toml"
	zvpTxt     = currDir + "ZvP.txt"
	zvtTxt     = currDir + "ZvT.txt"
	zvzTxt     = currDir + "ZvZ.txt"
	mmrDiffTxt = currDir + "MMR-diff.txt"
	winrateTxt = currDir + "winrate.txt"
	cfg        settings
)

// the cfg.toml file
type settings struct {
	mainToon, dir              string
	updateTime                 int64
	useAPI                     bool
	apiClientID, apiClientPass string
}

// fullPath is your cfg.toml file
func setup(absolutePath string) *player {
	pl := &player{
		[2]uint8{0, 0}, [2]uint8{0, 0}, [2]uint8{0, 0},
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
		apiClientID := config.Get("settings.apiClientID").(string)
		apiClientPass := config.Get("settings.apiClientPass").(string)

		for i := range toons {
			arr := toons[i].([]interface{})

			url := arr[0].(string)
			name := arr[1].(string)
			race := arr[2].(string)

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

			pl.profile[name] = profile
		}

		cfg = settings{
			mainToon, dir,
			updateTime,
			useAPI,
			apiClientID, apiClientPass,
		}
	} else {
		writeData(cfgToml, myToml)
		fmt.Println("Now setup your cfg.toml file.")
		os.Exit(0)
	}

	return pl
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

var myToml = `#     name - Put a comma-separated list of your SC2 accounts like in the example (url, name, race).
# mainToon - You must choose only one name to use.
#      dir - Where to watch for new SC2 replays (use either a single slash, or a double backslash).

[account]
name = [ [ "https://starcraft2.com/en-gb/profile/1/1/1331332", "Gixxasaurus", "zerg" ],
         [ "https://starcraft2.com/en-gb/profile/2/1/4545534", "Rairden", "zerg" ],
         [ "https://starcraft2.com/en-gb/profile/1/1/6901550", "PREAHLANY", "zerg"] ]

mainToon = "Gixxasaurus"

[directory]
dir = "/home/erik/scratch/replays/"
# dir = "C:/Users/Erik/Downloads/reps/"
# dir = "C:\\Users\\Erik\\Downloads\\reps\\"

[settings]
updateTime = 1000
useAPI = false
apiClientID = "632b0e2b3f0a4d64abf4060794fca015"
apiClientPass = "eR5qWtmpyzM4OWzRHqXzhkCwokOq8rEI"
`
