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
	dir string
	names []string
	player  = Player{[2]uint8{0, 0}, [2]uint8{0, 0}, [2]uint8{0, 0}}
	matchup = NIL
)

const (
	NIL = uint8(iota)
	ZvP
	ZvT
	ZvZ
)

type Player struct {
	ZvP [2]uint8
	ZvT [2]uint8
	ZvZ [2]uint8
}

func main() {
	dir = cfg.dir
	names = cfg.names
	files, _ := ioutil.ReadDir(dir)
	player.updateAllScores(files)
	saveAllFiles()

	fileCnt := numFiles(files)

	for {
		time.Sleep(1 * time.Second)

		if fileCnt == numFiles(files) {
			files, _ = ioutil.ReadDir(dir)
			continue
		}

		fileCnt = numFiles(files)
		lastModified := getLastModified(dir)
		player.updateScore(lastModified)

		saveFile()
	}
}

func (p *Player) updateAllScores(files []os.FileInfo) {
	for _, file := range files {
		p.updateScore(file)
	}
}

func (p *Player) updateScore(file os.FileInfo) {
	r, err := rep.NewFromFileEvts(dir + file.Name(), false, false, false)

	if err != nil {
		fmt.Printf("Failed to open file: '%v'\n", err)
		return
	}

	defer r.Close()
	parseReplay(r)
}

func parseReplay(r *rep.Rep) {
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
	var match bool
	for _, toon := range names {
		if *name == toon {
			match = true
			break
		}
	}

	if match {
		ZvX[0]++
	} else {
		ZvX[1]++
	}
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

func saveAllFiles() {
	writeFile(currDir + "ZvP.txt", &player.ZvP)
	writeFile(currDir + "ZvT.txt", &player.ZvT)
	writeFile(currDir + "ZvZ.txt", &player.ZvZ)
}

func writeFile(fullPath string, mu *[2]uint8) {
	file, e := os.Create(fullPath)
	check(e)
	defer file.Close()
	file.WriteString(matchupToString(mu))
	file.Sync()
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

func numFiles(files []os.FileInfo) int {
	return len(files)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
