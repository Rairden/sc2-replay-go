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
	_, r, _, _  = runtime.Caller(0) // linux (testing)
	s, _        = os.Executable()   // win10
	currDir     = path.Dir(s) + filepath.Join("/")
	cfg_toml    = currDir + "cfg.toml"
	ZvP_txt     = currDir + "ZvP.txt"
	ZvT_txt     = currDir + "ZvT.txt"
	ZvZ_txt     = currDir + "ZvZ.txt"
	MMRdiff_txt = currDir + "MMR-diff.txt"
	winrate_txt = currDir + "winrate.txt"
	cfg         settings
)

// the cfg.toml file
type settings struct {
	names                           []string
	dir, apiClientId, apiClientPass string
	useAPI                          bool
	updateTime, main                int64
}

func init() {
	if cfgExists() {
		config, _ := toml.Load(cfgToString())
		toons := config.Get("account.name").([]interface{})
		main := config.Get("account.main").(int64)
		dir := config.Get("directory.dir").(string)
		useAPI := config.Get("settings.useAPI").(bool)
		updateTime := config.Get("settings.updateTime").(int64)
		apiClientId := config.Get("settings.apiClientId").(string)
		apiClientPass := config.Get("settings.apiClientPass").(string)

		names := make([]string, len(toons))
		for i := range toons {
			arr := toons[i].([]interface{})

			url := arr[0].(string)
			playerName := arr[1].(string)
			race := arr[2].(string)
			names[i] = playerName

			split := strings.Split(url, "/")

			regionId := split[5]
			realmId := split[6]
			profileId := split[7]

			regionString := convertRegionToString(regionId)

			profile := &Profile{
				url, playerName, race,
				regionId, realmId, profileId,
				"", regionString,
			}

			player.profile = append(player.profile, *profile)
		}
		cfg = settings{
			names,
			dir, apiClientId, apiClientPass,
			useAPI,
			updateTime, main}
		fmt.Println()

	} else {
		writeData(cfg_toml, myToml)
		fmt.Println("Now setup your cfg.toml file.")
		os.Exit(0)
	}
}

func convertRegionToString(reg string) string {
	switch reg {
	case "1":
		return "us"
	case "2":
		return "eu"
	}
	return ""
}

func cfgExists() bool {
	_, err := os.Open(cfg_toml)

	if err != nil {
		return false
	}

	return true
}

func cfgToString() string {
	b, err := ioutil.ReadFile(cfg_toml)

	if err != nil {
		fmt.Printf("File not found: '%v'", b)
	}
	return string(b)
}

var myToml = `# name - Put a comma-separated list of your SC2 account like in the example (url, name, race).
# main - You have to choose ONLY one of your names (accounts). Counting starts at 0.
#  dir - Where to watch for new SC2 replays (use either a single slash, or a double backslash).

[account]
name = [ [ "https://starcraft2.com/en-gb/profile/1/1/1331332", "Gixxasaurus", "zerg" ],
         [ "https://starcraft2.com/en-gb/profile/2/1/4545534", "Rairden", "zerg" ] ]

main = 0

[directory]
dir = "/home/erik/scratch/replays/"
# dir = "C:/Users/Erik/Downloads/reps/"
# dir = "C:\\Users\\Erik\\Downloads\\reps\\"

[settings]
updateTime = 1000
useAPI = true
apiClientId = "632b0e2b3f0a4d64abf4060794fca015"
apiClientPass = "eR5qWtmpyzM4OWzRHqXzhkCwokOq8rEI"
`
