package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
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
	names []string
	dir   string
}

func init() {
	if cfgExists() {
		config, _ := toml.Load(cfgToString())
		toons := config.Get("account.name").([]interface{})
		dir := config.Get("directory.dir").(string)

		names := make([]string, len(toons))
		for i := range toons {
			arr := toons[i].([]interface{})

			url := arr[0].(string)
			playerName := arr[1].(string)
			race := arr[2].(string)
			names[i] = playerName

			split := strings.Split(url, "/")

			regionId, _ := strconv.Atoi(split[5])
			realmId, _ := strconv.Atoi(split[6])
			profileId, _ := strconv.Atoi(split[7])

			profile := &Profile{
				url, playerName, race, regionId, realmId, profileId, 0,
			}

			player.profile = append(player.profile, *profile)
		}
		cfg = settings{names, dir}

	} else {
		writeData(cfg_toml, config)
		fmt.Println("Now setup your cfg.toml file.")
		os.Exit(0)
	}
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

var config = `# name - Put a comma-separated list of your SC2 account like in example (url, name, race).
[account]
name = [ [ "https://starcraft2.com/en-gb/profile/1/1/1331332", "Gixxasaurus", "zerg" ],
		 [ "https://starcraft2.com/en-gb/profile/2/1/4545534", "Rairden", "zerg" ] ]
useAPI = true

# dir - Where to watch for new SC2 replays (use either a single slash, or a double backslash).
[directory]
#dir = "/home/erik/scratch/replays/"
#dir = "C:/Users/Erik/Downloads/reps/"
dir = "C:\\Users\\Erik\\Downloads\\reps\\"
`
