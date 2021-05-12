package main

import (
	"errors"
	"fmt"
	"github.com/icza/s2prot/rep"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

type player struct {
	ZvP, ZvT, ZvZ, total [2]uint8
	startMMR, MMR        int64
	profile              map[string]*profile
}

type player2 struct {
	player
	client *http.Client
}

type profile struct {
	url, name, race                        string
	regionID, realmID, profileID, ladderID string
	region                                 string
}

type user interface {
	getMMR() int64
}

func (p *player) getMMR() int64 {
	newestFile, err := getLastModified(cfg.dir)
	if err != nil {
		return 0
	}
	lastGame := fileToGame(newestFile)
	return p.getReplayMMR(lastGame)
}

func (p *player2) getMMR() int64 {
	return p.getMmrAPI(p.client)
}

// Both ranked/unranked has competitive = true, while an A.I. game is false.
type game struct {
	players       []toon
	matchup       string
	isCompetitive bool
}

type toon struct {
	profileID string
	name      string
	mmr       int64
	result    string
}

func main() {
	player := setup(cfgToml)
	player2 := &player2{*player, nil}

	fmt.Printf("Checking the directory '%v' \nevery %v ms for new SC2 replays...\n\n", cfg.dir, cfg.updateTime)

	if cfg.useAPI {
		mainAPI(player2)
	} else {
		mainNoAPI(player)
	}
}

func mainAPI(pl *player2) {
	files, _ := ioutil.ReadDir(cfg.dir) // TODO: replace with os.ReadDir

	if cfg.clientID == "" {
		err := getCredentials()
		if err != nil {
			fmt.Println(err)
			mainNoAPI(&pl.player)
		}
	}

	client := getBattleNetClient(cfg.clientID, cfg.clientSecret)
	pl.client = client
	pl.setLadderID(client) // 1) make request to ladder summary API. Get ladderID.

	// Allow user to start with a non-empty replay folder
	if len(files) > 0 {
		pl.setStartMMR(files)
		pl.updateAllScores(files)
		pl.writeMMRdiff(pl.startMMR, pl.MMR)
		pl.writeWinRate()
	} else {
		pl.MMR = pl.getMmrAPI(client)
		pl.startMMR = pl.MMR
		saveResetStats()
	}

	pl.saveAllFiles()
	pl.run(pl)
}

func mainNoAPI(pl *player) {
	files, _ := ioutil.ReadDir(cfg.dir)

	// Allow user to start with a non-empty replay folder
	if len(files) > 0 {
		pl.setStartMMR(files)
		pl.updateAllScores(files)
		newestFile, _ := getLastModified(cfg.dir)
		mmr := pl.getReplayMMR(fileToGame(newestFile))
		pl.writeMMRdiff(pl.startMMR, mmr)
		pl.writeWinRate()
	} else {
		saveResetStats()
	}

	pl.saveAllFiles()
	pl.run(pl)
}

func (p *player) run(usr user) {
	files, _ := os.ReadDir(cfg.dir)
	fileCnt := numFiles(files)
	fmt.Printf("Start MMR: %11v\n", p.startMMR)

	for {
		time.Sleep(time.Duration(cfg.updateTime) * time.Millisecond)

		if nf := numFiles(files); nf <= fileCnt {
			files, _ = os.ReadDir(cfg.dir)
			fileCnt = nf
			continue
		}

		fileCnt = numFiles(files)

		// If you don't want to restart program, you can just delete all replays from directory.
		if fileCnt == 0 {
			p.resetPlayer()
			p.saveAllFiles()
			saveResetStats()
			p.startMMR = usr.getMMR()
			continue
		}

		game := p.updateScore()
		if !game.isCompetitive {
			continue
		}

		p.writeWinRate()
		p.setTotalWinLoss()
		p.writeTotalWinLoss()
		p.saveFile(game.matchup)
		p.printResults(game)

		if fileCnt == 1 || p.startMMR == 0 {
			files, _ := ioutil.ReadDir(cfg.dir)
			p.setStartMMR(files)
		}

		p.MMR = usr.getMMR()
		p.writeMMRdiff(p.startMMR, p.MMR)
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

	// Only InitData shows it's an A.I. (computer) match at 'x.InitData.Struct.gameDescription.gameOptions.competitive'
	isCompetitive := initData.GameDescription.GameOptions.CompetitiveOrRanked()

	mu := getMatchup(Matchup)
	var toons []toon

	for i := 0; i < 2; i++ {
		p1 := toon{
			strconv.FormatInt(players[i].Toon.ID(), 10),
			userInitDatas[i].Name(),
			userInitDatas[i].MMR(),
			players[i].Result().String(),
		}
		toons = append(toons, p1)
	}

	return game{toons, mu, isCompetitive}
}

func (p *player) printResults(g game) {
	for _, pl := range g.players {
		if _, ok := p.profile[pl.profileID]; ok {
			fmt.Printf("%s %-11s %6v %s\n", g.matchup, pl.name, pl.mmr, pl.result)
			return
		}
	}
}

func getMatchup(mu string) string {
	switch mu {
	case "ZvP", "PvZ":
		return "ZvP"
	case "ZvT", "TvZ":
		return "ZvT"
	case "ZvZ":
		return "ZvZ"
	default:
		return ""
	}
}

func getWinner(g game) toon {
	if g.players[0].result == "Victory" {
		return g.players[0]
	}
	return g.players[1]
}

func (p *player) updateAllScores(files []os.FileInfo) {
	for _, file := range files {
		g := fileToGame(file)
		if !g.isCompetitive {
			continue
		}
		winner := getWinner(g)
		p.setScore(winner.profileID, g.matchup)
	}
	p.setTotalWinLoss()
	p.writeTotalWinLoss()
}

func (p *player) updateScore() game {
	f, _ := getLastModified(cfg.dir)
	g := fileToGame(f)
	if !g.isCompetitive {
		return g
	}
	winner := getWinner(g)
	p.setScore(winner.profileID, g.matchup)
	return g
}

// setScore The ID must be the winner.
func (p *player) setScore(ID, matchup string) {
	switch matchup {
	case "ZvP":
		p.incScore(ID, &p.ZvP)
	case "ZvT":
		p.incScore(ID, &p.ZvT)
	case "ZvZ":
		p.incScore(ID, &p.ZvZ)
	}
}

func (p *player) incScore(ID string, ZvX *[2]uint8) {
	isYou := p.isMyID(ID)

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
	p.total = [2]uint8{0, 0}
	p.MMR, p.startMMR = 0, 0
}

func (p *player) isMyID(ID string) bool {
	if _, ok := p.profile[ID]; ok {
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
	writeFile(totalWinLossTxt, &p.total)
}

func saveResetStats() {
	writeData(mmrDiffTxt, "+0 MMR\n")
	writeData(winrateTxt, "0%\n")
	writeData(totalWinLossTxt, " 0 - 0\n")
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

// Sort by modification time in ascending order
func sortFilesModTime(files []fs.FileInfo) []os.FileInfo {
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
	return files
}

// using path is more expensive than a []fs.FileInfo param, but I need to refresh dir
func getLastModified(path string) (os.FileInfo, error) {
	files, _ := ioutil.ReadDir(path)
	if len(files) == 0 {
		return nil, errors.New("error: no files")
	}

	sort.Slice(files, func(i, j int) bool {
		return files[j].ModTime().Before(files[i].ModTime())
	})
	return files[0], nil
}

func numFiles(files []os.DirEntry) int {
	return len(files)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
