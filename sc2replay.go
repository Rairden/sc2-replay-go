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
	dir     string
	names   []string
	player  = Player{[2]uint8{0, 0}, [2]uint8{0, 0}, [2]uint8{0, 0}, 0, 0}
	matchup = NIL
)

const (
	NIL = uint8(iota)
	ZvP
	ZvT
	ZvZ
)

type Player struct {
	ZvP      [2]uint8
	ZvT      [2]uint8
	ZvZ      [2]uint8
	startMMR float64
	MMR      float64
}

func main() {
	dir = cfg.dir
	names = cfg.names
	files, _ := ioutil.ReadDir(dir)

	// set start MMR and current MMR (if starting w/ non-empty folder)
	if len(files) > 1 {
		oldestFile := getLeastModified(dir)
		newestFile := getLastModified(dir)
		firstRep := decodeReplay(oldestFile)
		player.startMMR = player.setMMR(firstRep)

		replay := decodeReplay(newestFile)
		player.setMMR(replay)

		player.updateAllScores(files)
		player.writeMMRdiff()
		player.writeWinRate()
		saveAllFiles(true)
	} else {
		saveAllFiles(false)
	}

	fileCnt := numFiles(files)

	for {
		time.Sleep(1 * time.Second)

		if fileCnt == numFiles(files) {
			files, _ = ioutil.ReadDir(dir)
			continue
		}

		fileCnt = numFiles(files)
		lastModified := getLastModified(dir)

		if numFiles(files) == 0 || lastModified == nil {
			player.resetScores()
			saveAllFiles(false)
			continue
		}

		player.parseReplay(lastModified)
		replay := decodeReplay(lastModified)
		player.setMMR(replay)
		saveFile()
		player.writeWinRate()

		if numFiles(files) == 1 {
			mmr := player.setMMR(replay)
			player.startMMR = mmr
		}
		player.writeMMRdiff()
	}
}

func decodeReplay(file os.FileInfo) *rep.Rep {
	r, err := rep.NewFromFileEvts(dir + file.Name(), false, false, false)
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
	isPlayer := checkIfPlayer(name)

	if isPlayer {
		ZvX[0]++
	} else {
		ZvX[1]++
	}
}

func checkIfPlayer(name *string) bool {
	var match bool
	for _, toon := range names {
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
		writeFile(currDir + "ZvP.txt", &player.ZvP)
	case ZvT:
		writeFile(currDir + "ZvT.txt", &player.ZvT)
	case ZvZ:
		writeFile(currDir + "ZvZ.txt", &player.ZvZ)
	}
}

func saveAllFiles(skipMMR bool) {
	writeFile(currDir + "ZvP.txt", &player.ZvP)
	writeFile(currDir + "ZvT.txt", &player.ZvT)
	writeFile(currDir + "ZvZ.txt", &player.ZvZ)

	if !skipMMR {
		writeData(currDir + "MMR-diff.txt", "+0 MMR\n")
		writeData(currDir + "winrate.txt", "0%\n")
	}
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

	if len(files) == 0 {
		return nil
	}

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

func (p *Player) resetScores() {
	p.MMR = 0
	p.startMMR = 0

	p.ZvP[0] = 0
	p.ZvT[0] = 0
	p.ZvZ[0] = 0
	p.ZvP[1] = 0
	p.ZvT[1] = 0
	p.ZvZ[1] = 0
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
