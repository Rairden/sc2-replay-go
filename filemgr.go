package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

var (
	_, s, _, _ = runtime.Caller(0)
	currDir    = filepath.Dir(s) + filepath.Join("/")

	cfg settings
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
			names[i] = toons[i].(string)
		}
		cfg = settings{names, dir}

	} else {
		writeData(currDir + "cfg.toml", config)
		fmt.Println("Now setup your cfg.toml file.")
		os.Exit(0)
	}
}

func cfgExists() bool {
	cfg := currDir + filepath.Join("cfg.toml")
	fmt.Printf("    cfg: %v\n", cfg)
	_, err := os.Open(cfg)

	if err != nil {
		return false
	}

	return true
}

func cfgToString() string {
	fmt.Printf("currDir: %v\n", currDir)
	b, err := ioutil.ReadFile(currDir + "cfg.toml")

	if err != nil {
		fmt.Printf("File not found: '%v'", b)
	}

	str := string(b)
	return str
}

var config = `# name - Put a comma-separated list of your SC2 player names. ID not required.
[account]
name = [ "Gixxasaurus", "Rairden" ]
ID = [ 1331332, 4545534 ]

# dir - Where to watch for new SC2 replays
[directory]
dir = "/home/erik/scratch/replays/"
`
