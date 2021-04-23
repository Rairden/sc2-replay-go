package main

import (
	"fmt"
	"github.com/icza/s2prot/rep"
	"io/fs"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"time"
)

type player struct {
	ZvP, ZvT, ZvZ [2]uint8
	startMMR, MMR int64
	profile       map[string]*profile
}

type profile struct {
	url, name, race                        string
	regionID, realmID, profileID, ladderID string
	region                                 string
}

type game struct {
	players []toon
	matchup string
}

type toon struct {
	profileID int64
	name      string
	mmr       int64
	result    string
}

func main() {
	player := setup(cfgToml)

	fmt.Printf("Checking the directory '%v' every %v milliseconds for new SC2 replays...\n", cfg.dir, cfg.updateTime)

	if cfg.useAPI {
		mainAPI(player)
	} else {
		mainNoAPI(player)
	}
}

func mainAPI(pl *player) {
	files, _ := ioutil.ReadDir(cfg.dir)
	client := getBattleNetClient()
	pl.setLadderID(client) // 1. make request to ladder summary API. Get ladderID.

	var apiMMR int64

	if pl.profile[cfg.mainToon].ladderID != "" {
		apiMMR = int64(pl.getMmrAPI(client))
		if apiMMR != 0 {
			pl.startMMR = apiMMR
		}
	}

	// set start MMR and current MMR (if starting w/ non-empty folder)
	if len(files) >= 1 {
		oldestFile := getFirstModified(cfg.dir)
		firstGame := fileToGame(oldestFile)
		pl.startMMR = pl.setMMR(firstGame)

		pl.updateAllScores(files)

		if pl.MMR != 0 {
			pl.calcMMRdiffAPI(apiMMR)
		}

		pl.writeWinRate()
		pl.saveAllFiles()
	} else {
		pl.saveAllFiles()
		saveResetStats()
	}

	fileCnt := numFiles(files)

	for {
		time.Sleep(time.Duration(cfg.updateTime) * time.Millisecond)

		if fileCnt == numFiles(files) {
			files, _ = ioutil.ReadDir(cfg.dir)
			continue
		}

		fileCnt = numFiles(files)

		// If you don't want to restart program, you can just delete all replays from directory.
		if numFiles(files) == 0 {
			pl.resetPlayer()
			apiMMR = int64(pl.getMmrAPI(client))
			pl.startMMR = apiMMR
			pl.saveAllFiles()
			saveResetStats()
			continue
		}

		newestFile := getLastModified(cfg.dir)
		game := fileToGame(newestFile)
		pl.SetScore(game.players[0].name, game.matchup)

		mmr := pl.getMmrAPI(client)
		pl.MMR = int64(mmr)

		if apiMMR != 0 {
			pl.calcMMRdiffAPI(int64(mmr))
		}
		pl.writeWinRate()
		pl.saveFile(game.matchup)
	}
}

func mainNoAPI(pl *player) {
	files, _ := ioutil.ReadDir(cfg.dir)
	// set start MMR and current MMR (if starting w/ non-empty folder)
	if len(files) >= 1 {
		pl.startMMR = pl.setStartMMR(files)

		newestFile := getLastModified(cfg.dir)
		newestGame := fileToGame(newestFile)
		pl.setMMR(newestGame)

		pl.updateAllScores(files)
		pl.writeMMRdiff()
		pl.writeWinRate()
		pl.saveAllFiles()
	} else {
		pl.saveAllFiles()
		saveResetStats()
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
			pl.resetPlayer()
			pl.saveAllFiles()
			saveResetStats()
			continue
		}

		game := pl.updateScore()
		pl.setMMR(game)
		pl.writeMMRdiff()
		pl.writeWinRate()
		pl.saveFile(game.matchup)
	}
}

func decodeReplay(file os.FileInfo) *rep.Rep {
	r, err := rep.NewFromFileEvts(cfg.dir+file.Name(), false, false, false)
	check(err)
	defer r.Close()
	return r
}

func fileToGame(file fs.FileInfo) game {
	replay := decodeReplay(file)
	return getGame(replay)
}

func getInitData(r *rep.Rep) *rep.InitData {
	return &r.InitData
}

// toon has 4 fields. Three from rep.Details, and one from rep.InitData because the name is unreliable from Details.
func getGame(r *rep.Rep) game {
	Matchup := r.Details.Matchup()
	players := r.Details.Players()
	initData := getInitData(r)
	userInitDatas := initData.UserInitDatas

	mu := getMatchup(Matchup)
	var toons []toon

	for i := 0; i < 2; i++ {
		p1 := toon{
			players[i].Toon.ID(),
			userInitDatas[i].Name(),
			userInitDatas[i].MMR(),
			players[i].Result().String(),
		}
		toons = append(toons, p1)
	}

	return game{toons, mu}
}

func (p *player) printResults(g game) {
	for _, pl := range g.players {
		if _, ok := p.profile[pl.name]; ok {
			fmt.Printf("%s %-11s %6v %s\n", g.matchup, pl.name, pl.mmr, pl.result)
			return
		}
	}
}

func getMatchup(mu string) string {
	if mu == "PvZ" || mu == "ZvP" {
		return "ZvP"
	}
	if mu == "TvZ" || mu == "ZvT" {
		return "ZvT"
	}
	if mu == "ZvZ" {
		return "ZvZ"
	}
	return ""
}

// FIXME: consider changing this to use your toon and not grabbing winner.
func (p *player) updateAllScores(files []os.FileInfo) {
	for _, file := range files {
		g := fileToGame(file)
		if g.players[0].result == "Victory" {
			p.SetScore(g.players[0].name, g.matchup)
		} else {
			p.SetScore(g.players[1].name, g.matchup)
		}
	}
}

func (p *player) updateScore() game {
	f := getLastModified(cfg.dir)
	g := fileToGame(f)
	p.SetScore(g.players[0].name, g.matchup)
	return g
}

// SetScore The name can be winner or loser.
func (p *player) SetScore(name, matchup string) {
	switch matchup {
	case "ZvP":
		p.incScore(name, &p.ZvP)
	case "ZvT":
		p.incScore(name, &p.ZvT)
	case "ZvZ":
		p.incScore(name, &p.ZvZ)
	}
}

func (p *player) incScore(name string, ZvX *[2]uint8) {
	isYou := p.isMyName(name)

	if isYou {
		ZvX[0]++
	} else {
		ZvX[1]++
	}
}

func (p *player) resetPlayer() {
	p.ZvP = [2]uint8{0, 0}
	p.ZvT = [2]uint8{0, 0}
	p.ZvZ = [2]uint8{0, 0}
	p.MMR, p.startMMR = 0, 0
}

func (p *player) isMyName(name string) bool {
	if _, ok := p.profile[name]; ok {
		return true
	}
	return false
}

func (p *player) saveFile(matchup string) {
	switch matchup {
	case "ZvP":
		writeFile(zvpTxt, &p.ZvP)
	case "ZvT":
		writeFile(zvtTxt, &p.ZvT)
	case "ZvZ":
		writeFile(zvzTxt, &p.ZvZ)
	}
}

func (p *player) saveAllFiles() {
	writeFile(zvpTxt, &p.ZvP)
	writeFile(zvtTxt, &p.ZvT)
	writeFile(zvzTxt, &p.ZvZ)
}

func saveResetStats() {
	writeData(mmrDiffTxt, "+0 MMR\n")
	writeData(winrateTxt, "0%\n")
}

func writeFile(fullPath string, mu *[2]uint8) {
	writeData(fullPath, scoreToString(mu))
}

func scoreToString(ZvX *[2]uint8) string {
	win := strconv.Itoa(int(ZvX[0]))
	loss := strconv.Itoa(int(ZvX[1]))
	str := fmt.Sprintf("%2s - %s\n", win, loss)
	return str
}

// sort by modification time ascending
func sortFilesModTime(files []fs.FileInfo) []os.FileInfo {
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
	return files
}

// using path is more expensive than a []fs.FileInfo param, but I need to refresh dir
func getLastModified(path string) os.FileInfo {
	files, _ := ioutil.ReadDir(path)

	sort.Slice(files, func(i, j int) bool {
		return files[j].ModTime().Before(files[i].ModTime())
	})
	return files[0]
}

func getFirstModified(path string) os.FileInfo {
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
