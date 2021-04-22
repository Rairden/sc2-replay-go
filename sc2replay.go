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

var player = &Player{
	[2]uint8{0, 0}, [2]uint8{0, 0}, [2]uint8{0, 0},
	0, 0,
	make(map[string]*Profile),
}

const (
	NIL uint8 = iota
	ZvP
	ZvT
	ZvZ
)

type Game struct {
	players []toon
	matchup uint8
}

type Player struct {
	ZvP, ZvT, ZvZ [2]uint8
	startMMR, MMR int64
	profile       map[string]*Profile
}

type Profile struct {
	url, name, race                        string
	regionId, realmId, profileId, ladderId string
	region                                 string
}

type toon struct {
	profileId int64
	name      string
	mmr       int64
	result    string
	matchup   uint8
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

	var apiMMR int64

	if player.profile[cfg.main].ladderId != "" {
		apiMMR = int64(getMMR(client))
		if apiMMR != 0 {
			player.startMMR = apiMMR
		}
	}

	// set start MMR and current MMR (if starting w/ non-empty folder)
	if len(files) >= 1 {
		oldestFile := getLeastModified(cfg.dir)
		firstGame := fileToGame(oldestFile)
		player.startMMR = player.setMMR(firstGame)

		player.updateAllScores(files)

		if player.MMR != 0 {
			player.calcMMRdiffAPI(apiMMR)
		}

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
			apiMMR = int64(getMMR(client))
			player.startMMR = apiMMR
			saveAllFiles()
			saveResetMMR()
			continue
		}

		newestFile := getLastModified(cfg.dir)
		game := fileToGame(newestFile)
		player.SetScore(game.players[0].name, game.matchup)

		mmr := getMMR(client)
		player.MMR = int64(mmr)

		if apiMMR != 0 {
			player.calcMMRdiffAPI(int64(mmr))
		}
		player.writeWinRate()
		saveFile(game.matchup)
	}
}

func mainNoAPI() {
	files, _ := ioutil.ReadDir(cfg.dir)
	// set start MMR and current MMR (if starting w/ non-empty folder)
	if len(files) >= 1 {
		oldestFile := getLeastModified(cfg.dir)
		firstGame := fileToGame(oldestFile)
		player.startMMR = player.setMMR(firstGame)

		newestFile := getLastModified(cfg.dir)
		lastGame := fileToGame(newestFile)
		player.setMMR(lastGame)

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
			player.resetPlayer()
			saveAllFiles()
			saveResetMMR()
			continue
		}

		newestFile := getLastModified(cfg.dir)
		game := fileToGame(newestFile)
		player.SetScore(game.players[0].name, game.matchup)
		mmr := player.setMMR(game)
		player.writeMMRdiff()
		player.writeWinRate()
		saveFile(game.matchup)

		if numFiles(files) == 1 || player.startMMR == 0 {
			player.startMMR = mmr
		}
	}
}

func setLadderId(client *http.Client) {
	// set ladderId if not set
	if player.profile[cfg.main].ladderId == "" {
		ladderSummaryAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/summary?locale=en_US",
			player.profile[cfg.main].region, player.profile[cfg.main].regionId,
			player.profile[cfg.main].realmId, player.profile[cfg.main].profileId)

		apiLadderId := getLadderSummary(client, ladderSummaryAPI, player.profile[cfg.main].race)
		player.profile[cfg.main].ladderId = apiLadderId
	}
}

func getMMR(client *http.Client) int {
	// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/298683?locale=en_US&access_token=xxx
	ladderAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/%s?locale=en_US",
		player.profile[cfg.main].region, player.profile[cfg.main].regionId,
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

func (p *Player) updateAllScores(files []os.FileInfo) {
	for _, file := range files {
		replay := decodeReplay(file)
		game := getGame(replay)
		if game.players[0].result == "Victory" {
			p.SetScore(game.players[0].name, game.matchup)
		} else {
			p.SetScore(game.players[1].name, game.matchup)
		}
	}
}

func fileToGame(file os.FileInfo) Game {
	replay := decodeReplay(file)
	return getGame(replay)
}

func getGame(r *rep.Rep) Game {
	Matchup := r.Details.Matchup()
	players := r.Details.Players()
	initData := getInitData(r)
	userInitDatas := initData.UserInitDatas

	mu := getMatchup(Matchup)
	var toons []toon

	for i := 0; i < 2; i++ {
		p1 := toon{
			name:      userInitDatas[i].Name(),
			mmr:       userInitDatas[i].MMR(),
			profileId: players[i].Toon.ID(),
			result:    players[i].Result().String(),
			matchup: mu,
		}
		toons = append(toons, p1)
	}

	for i := 0; i < 2; i++ {
		player.printResults(i, toons)
	}

	return Game{players: toons, matchup: mu}
}

func (p *Player) printResults(i int, toons []toon) {
	if _, ok := p.profile[toons[i].name]; ok {
		fmt.Printf("%-20v %8v  %-8v %v\n", toons[i].name, toons[i].mmr, toons[i].result, toons[i].matchup)
	}
}

func getInitData(r *rep.Rep) *rep.InitData {
	return &r.InitData
}

func getMatchup(mu string) uint8 {
	if mu == "PvZ" || mu == "ZvP" {
		return ZvP
	}
	if mu == "TvZ" || mu == "ZvT" {
		return ZvT
	}
	if mu == "ZvZ" {
		return ZvZ
	}
	return NIL
}

// SetScore give this any name, and matchup (0 NIL, 1 ZvP, 2 ZvT, 3 ZvZ)
func (p *Player) SetScore(name string, matchup uint8) {
	switch matchup {
	case ZvP:
		incScore(name, &p.ZvP)
	case ZvT:
		incScore(name, &p.ZvT)
	case ZvZ:
		incScore(name, &p.ZvZ)
	}
}

func incScore(name string, ZvX *[2]uint8) {
	isPlayer := isMyName(name)

	if isPlayer {
		ZvX[0]++
	} else {
		ZvX[1]++
	}
}

func (p *Player) resetPlayer() {
	p.ZvP = [2]uint8{0, 0}
	p.ZvT = [2]uint8{0, 0}
	p.ZvZ = [2]uint8{0, 0}
	p.MMR, p.startMMR = 0, 0
}

func isMyName(name string) bool {
	for _, toon := range cfg.names {
		if name == toon {
			return true
		}
	}
	return false
}

func saveFile(matchup uint8) {
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
