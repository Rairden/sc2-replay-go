package main

import (
	"context"
	"fmt"
	"github.com/icza/s2prot/rep"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
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
	startMMR, MMR int
	profile       []Profile
}

type Profile struct {
	url, name, race                        string
	regionId, realmId, profileId, ladderId string
	regionString                           string
}

func main() {
	fmt.Printf("Checking the directory '%v' every %v milliseconds for new SC2 replays...\n", cfg.dir, cfg.updateTime)

	if cfg.useAPI {
		mainAPI()
	} else {
		mainNoAPI()
	}
}

func mainAPI() {
	files, _ := ioutil.ReadDir(cfg.dir)
	client := getBattleNetClient()
	setLadderId(client)		// 1. make request to ladder summary API. Get ladderId.

	currMMR := 0
	if player.profile[cfg.main].ladderId != "" {
		currMMR = getMMR(client)
		if currMMR != 0 {
			player.startMMR = currMMR
		}
	}

	// set start MMR and current MMR (if starting w/ non-empty folder)
	if len(files) >= 1 {
		oldestFile := getLeastModified(cfg.dir)
		firstRep := decodeReplay(oldestFile)
		player.startMMR = player.setMMR(firstRep)

		updateAllScores(files)
		player.calcMMRdiffAPI(currMMR)

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

		// If you don't want to restart program, you can just delete all replays from directory.
		if numFiles(files) == 0 {
			player.resetPlayer()
			currMMR = getMMR(client)
			if currMMR != 0 {
				player.startMMR = currMMR
			}
			saveAllFiles()
			saveResetMMR()
			continue
		}

		lastModified := getLastModified(cfg.dir)
		parseReplay(lastModified)
		replay := decodeReplay(lastModified)
		_ = replay

		player.calcMMRdiffAPI(getMMR(client))
		player.writeWinRate()
		saveFile()

		// todo: fix me (api NO)
		if numFiles(files) == 1 || player.startMMR == 0 {
			// player.startMMR = mmr
		}
	}
}

func (p *Player) resetPlayer() {
	p.ZvP = [2]uint8{0, 0}
	p.ZvT = [2]uint8{0, 0}
	p.ZvZ = [2]uint8{0, 0}
	p.MMR, p.startMMR = 0, 0
}

func mainNoAPI() {
	files, _ := ioutil.ReadDir(cfg.dir)
	// set start MMR and current MMR (if starting w/ non-empty folder)
	if len(files) >= 1 {
		oldestFile := getLeastModified(cfg.dir)
		newestFile := getLastModified(cfg.dir)
		firstRep := decodeReplay(oldestFile)
		player.startMMR = player.setMMR(firstRep)

		replay := decodeReplay(newestFile)
		player.setMMR(replay)

		updateAllScores(files)
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
			player.resetPlayer()
			saveAllFiles()
			saveResetMMR()
			continue
		}

		lastModified := getLastModified(cfg.dir)
		parseReplay(lastModified)
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

func setLadderId(client *http.Client) {
	// set ladderId if not set
	if player.profile[cfg.main].ladderId == "" {
		ladderSummaryAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/summary?locale=en_US",
			player.profile[cfg.main].regionString, player.profile[cfg.main].regionId,
			player.profile[cfg.main].realmId, player.profile[cfg.main].profileId)

		player.profile[cfg.main].ladderId = getLadderSummary(client, ladderSummaryAPI, player.profile[cfg.main].race)
	}
}

func getMMR(client *http.Client) int {
	// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/298683?locale=en_US&access_token=xxx
	ladderAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/%s?locale=en_US",
		player.profile[cfg.main].regionString, player.profile[cfg.main].regionId,
		player.profile[cfg.main].realmId, player.profile[cfg.main].profileId, player.profile[cfg.main].ladderId)

	return getLadder(client, ladderAPI)
}

func getBattleNetClient() *http.Client {
	config := &clientcredentials.Config{
		ClientID:     cfg.apiClientId,
		ClientSecret: cfg.apiClientPass,
		TokenURL:     "https://us.battle.net/oauth/token",
	}

	// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/summary?locale=en_US&access_token=xxx
	client := config.Client(context.Background())
	return client
}

func decodeReplay(file os.FileInfo) *rep.Rep {
	r, err := rep.NewFromFileEvts(cfg.dir+file.Name(), false, false, false)
	check(err)
	defer r.Close()
	return r
}

func parseReplay(file os.FileInfo) {
	r := decodeReplay(file)
	updateScore(r)
}

func updateAllScores(files []os.FileInfo) {
	for _, file := range files {
		parseReplay(file)
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

// todo: fix/replace updateScore(*rep.Rep) so it also pulls from InitData (name and optionally MMR)
func foo(r *rep.Rep) {
	type toon struct {
		Name string
		mmr int64
	}

	// var players []toon
	// initData := r.InitData.UserInitDatas
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
	isPlayer := isMyName(name)

	if isPlayer {
		ZvX[0]++
	} else {
		ZvX[1]++
	}
}

func isMyName(name *string) bool {
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
	writeData(fullPath, scoreToString(mu))
}

func scoreToString(ZvX *[2]uint8) string {
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
