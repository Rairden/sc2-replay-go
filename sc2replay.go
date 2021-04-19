package main

import (
	"fmt"
	"github.com/icza/s2prot/rep"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"time"
)

var (
	player  = Player{[2]uint8{0, 0}, [2]uint8{0, 0}, [2]uint8{0, 0}, 0, 0, nil}
	matchup = NIL
)

const (
	NIL = uint8(iota)
	ZvP
	ZvT
	ZvZ
)

type Player struct {
	ZvP, ZvT, ZvZ [2]uint8
	startMMR, MMR float64
	profile       []Profile
}

type Profile struct {
	url, name, race                        string
	regionId, realmId, profileId, ladderId int
}

func main() {
	fmt.Printf("Checking the directory '%v' every %v milliseconds for new SC2 replays...\n", cfg.dir, cfg.updateTime)
	files, _ := ioutil.ReadDir(cfg.dir)

	// set start MMR and current MMR (if starting w/ non-empty folder)
	if len(files) >= 1 {
		oldestFile := getLeastModified(cfg.dir)
		newestFile := getLastModified(cfg.dir)
		firstRep := decodeReplay(oldestFile)
		player.startMMR = player.setMMR(firstRep)

		replay := decodeReplay(newestFile)
		player.setMMR(replay)

		player.updateAllScores(files)
		player.writeMMRdiff()
		player.writeWinRate()
		saveAllFiles()
	} else {
		saveAllFiles()
		saveResetMMR()
	}

	fileCnt := numFiles(files)

	for {
		time.Sleep(time.Duration(cfg.updateTime) * time.Millisecond)

		if fileCnt == numFiles(files) {
			files, _ = ioutil.ReadDir(cfg.dir)
			continue
		}

		fileCnt = numFiles(files)

		if numFiles(files) == 0 {
			player = Player{}
			saveAllFiles()
			saveResetMMR()
			continue
		}

		lastModified := getLastModified(cfg.dir)
		player.parseReplay(lastModified)
		replay := decodeReplay(lastModified)
		mmr := player.setMMR(replay)
		player.writeMMRdiff()
		player.writeWinRate()
		saveFile()

		if numFiles(files) == 1 || player.startMMR == 0 {
			player.startMMR = mmr
		}
	}
}

func decodeReplay(file os.FileInfo) *rep.Rep {
	r, err := rep.NewFromFileEvts(cfg.dir+file.Name(), false, false, false)
	check(err)
	defer r.Close()
	return r
}

func (p *Player) parseReplay(file os.FileInfo) {
	r := decodeReplay(file)
	updateScore(r)
}

func (p *Player) updateAllScores(files []os.FileInfo) {
	for _, file := range files {
		p.parseReplay(file)
	}
}

func updateScore(r *rep.Rep) {
	matchup := r.Details.Matchup()
	players := r.Details.Players()

	setMatchup(&matchup)

	if players[0].Result().Name == "Victory" {
		player.SetScore(&players[0].Name)
	} else {
		player.SetScore(&players[1].Name)
	}
}

func setMatchup(mu *string) {
	if *mu == "PvZ" || *mu == "ZvP" {
		matchup = ZvP
		return
	}
	if *mu == "TvZ" || *mu == "ZvT" {
		matchup = ZvT
		return
	}
	if *mu == "ZvZ" {
		matchup = ZvZ
	}
}

func (p *Player) SetScore(name *string) {
	switch matchup {
	case ZvP:
		incScore(name, &p.ZvP)
	case ZvT:
		incScore(name, &p.ZvT)
	case ZvZ:
		incScore(name, &p.ZvZ)
	}
}

func incScore(name *string, ZvX *[2]uint8) {
	isPlayer := isPlayer(name)

	if isPlayer {
		ZvX[0]++
	} else {
		ZvX[1]++
	}
}

func isPlayer(name *string) bool {
	var match bool
	for _, toon := range cfg.names {
		if *name == toon {
			match = true
			break
		}
	}
	return match
}

func saveFile() {
	switch matchup {
	case ZvP:
		writeFile(ZvP_txt, &player.ZvP)
	case ZvT:
		writeFile(ZvT_txt, &player.ZvT)
	case ZvZ:
		writeFile(ZvZ_txt, &player.ZvZ)
	}
}

func saveAllFiles() {
	writeFile(ZvP_txt, &player.ZvP)
	writeFile(ZvT_txt, &player.ZvT)
	writeFile(ZvZ_txt, &player.ZvZ)
}

func saveResetMMR() {
	writeData(MMRdiff_txt, "+0 MMR\n")
	writeData(winrate_txt, "0%\n")
}

func writeFile(fullPath string, mu *[2]uint8) {
	writeData(fullPath, matchupToString(mu))
}

func matchupToString(ZvX *[2]uint8) string {
	win := strconv.Itoa(int(ZvX[0]))
	loss := strconv.Itoa(int(ZvX[1]))
	str := fmt.Sprintf("%2s - %s\n", win, loss)
	return str
}

func getLastModified(path string) os.FileInfo {
	files, _ := ioutil.ReadDir(path)

	sort.Slice(files, func(i, j int) bool {
		return files[j].ModTime().Before(files[i].ModTime())
	})
	return files[0]
}

func getLeastModified(path string) os.FileInfo {
	files, _ := ioutil.ReadDir(path)

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
	return files[0]
}

func numFiles(files []os.FileInfo) int {
	return len(files)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
