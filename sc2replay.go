package main

import (
	"fmt"
	"github.com/icza/s2prot/rep"
	"io/ioutil"
	"regexp"
)

// var regex = regexp.MustCompile("^(1331332|4545534)$")
var regex = regexp.MustCompile("^(Gixxasaurus|Rairden)$")

func main() {

	player := Player{[2]uint8{0, 0}, [2]uint8{0, 0}, [2]uint8{0, 0}}

	// dir := "/home/erik/scratch/replays2/"
	largeDir := "/run/media/erik/storage/SC2/replayBackup/"
	files, _ := ioutil.ReadDir(largeDir)

	for _, file := range files {
		r, err := rep.NewFromFileEvts(largeDir + file.Name(), false, false, false)
		// r, err := rep.NewFromFile(largeDir + file.Name())
		if err != nil {
			fmt.Printf("Failed to open file: '%v'\n", err)
			return
		}
		defer r.Close()

		// id := p.Struct.Value("toon", "id")		// requires fmt.Sprintf (cast interface to string/int)
		matchup := r.Details.Matchup()
		players := r.Details.Players()

		// bug - 33% of the time the matchup is backwards and lists zerg first. ZvT, TvZ
		if matchup == "ZvT" || matchup == "ZvP" {
			createMatchup(&matchup)
		}

		if players[0].Result().Name == "Victory" {
			player.SetScore(&matchup, &players[0].Name)
		} else {
			player.SetScore(&matchup, &players[1].Name)
		}

	}
	// fmt.Println(player.ZvP)
	// fmt.Println(player.ZvT)
	// fmt.Println(player.ZvZ)
}

type Player struct {
	ZvP [2]uint8
	ZvT [2]uint8
	ZvZ [2]uint8
}

func (p *Player) SetScore(matchup *string, name *string) {
	switch *matchup {
	case "PvZ":
		if regex.MatchString(*name) {
			p.ZvP[0]++
		} else {
			p.ZvP[1]++
		}
	case "TvZ":
		if regex.MatchString(*name) {
			p.ZvT[0]++
		} else {
			p.ZvT[1]++
		}
	case "ZvZ":
		if regex.MatchString(*name) {
			p.ZvZ[0]++
		} else {
			p.ZvZ[1]++
		}
	}
}

func createMatchup(matchup *string) {
	if *matchup == "ZvP" {
		*matchup = "PvZ"
	}
	if *matchup == "ZvT" {
		*matchup = "TvZ"
	}
}
