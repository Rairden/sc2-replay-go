package main

import (
	"errors"
	"fmt"
	"github.com/icza/s2prot/rep"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	getMMR(files ...[]os.DirEntry) (int64, error)
}

func (p *player) getMMR(files ...[]os.DirEntry) (int64, error) {
	newestFile, err := getLastModified(files[0])
	if err != nil {
		return 0, err
	}
	lastGame := fileToGame(newestFile)
	return p.getReplayMMR(lastGame)
}

func (p *player2) getMMR(files ...[]os.DirEntry) (int64, error) {
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
	files := getAllReplays(cfg.dir)

	if cfg.clientID == "" {
		err := getCredentials()
		if err != nil {
			redirectError(err)
			mainNoAPI(&pl.player)
		}
	}

	client := getBattleNetClient(cfg.clientID, cfg.clientSecret)
	pl.client = client
	err := pl.setLadderID(client) // 1) make request to ladder summary API. Get ladderID.
	if err != nil {
		redirectError(err)
		mainNoAPI(&pl.player)
	}

	// Allow user to start with a non-empty replay folder
	if len(files) > 0 {
		pl.setStartMMR(files)
		pl.updateAllScores(files)
		pl.writeMMRdiff(pl.startMMR, pl.MMR)
		pl.writeWinRate()
	} else {
		saveResetStats()
	}

	mmr, err := pl.getMmrAPI(client)
	if err != nil {
		redirectError(err)
		mainNoAPI(&pl.player)
	}

	if pl.startMMR == 0 {
		pl.startMMR = mmr
		pl.MMR = mmr
	}

	pl.saveAllFiles()
	pl.run(pl)
}

func mainNoAPI(pl *player) {
	files := getAllReplays(cfg.dir)

	// Allow user to start with a non-empty replay folder
	if len(files) > 0 {
		pl.setStartMMR(files)
		pl.updateAllScores(files)
		newestGame := getLastModifiedGame(files)
		mmr, _ := pl.getReplayMMR(newestGame)
		pl.writeMMRdiff(pl.startMMR, mmr, len(files))
		pl.writeWinRate()
	} else {
		saveResetStats()
	}

	pl.saveAllFiles()
	pl.run(pl)
}

func (p *player) run(usr user) {
	files := getAllReplays(cfg.dir)
	fileCnt := len(files)
	isFirstLoop := true
	fmt.Printf("%12v %6v\n", "Start MMR:", p.startMMR)

	for {
		time.Sleep(time.Duration(cfg.updateTime) * time.Millisecond)

		// ignore deleting files unless the number of files is 0
		if len(files) <= fileCnt {
			if !isFirstLoop && len(files) == 0 {
				p.resetStats(usr, files)
				isFirstLoop = true
				fileCnt = len(files)
				continue
			}
			files = getAllReplays(cfg.dir)
			isFirstLoop = false
			continue
		}

		isFirstLoop = false
		fileCnt = len(files)

		game, err := p.updateScore(files)
		if err != nil {
			continue
		}

		p.writeWinRate()
		p.setTotalWinLoss()
		p.writeTotalWinLoss()
		p.saveFile(game.matchup)

		if fileCnt == 1 || p.startMMR == 0 {
			p.setStartMMR(files)
		}

		p.MMR, _ = usr.getMMR(files)
		p.writeMMRdiff(p.startMMR, p.MMR, len(files))
		p.printResults(game)
	}
}

// If you don't want to restart program, you can just delete all replays from directory.
func (p *player) resetStats(usr user, files []os.DirEntry) {
	p.resetPlayer()
	p.saveAllFiles()
	saveResetStats()
	p.startMMR, _ = usr.getMMR(files)
}

func getAllReplays(fullpath string) []os.DirEntry {
	files, _ := os.ReadDir(fullpath)
	var replays []os.DirEntry

	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".SC2Replay" {
			replays = append(replays, f)
		}
	}
	return replays
}

func decodeReplay(file os.DirEntry) *rep.Rep {
	r, err := rep.NewFromFileEvts(cfg.dir+file.Name(), false, false, false)
	check(err)
	defer r.Close()
	return r
}

func fileToGame(file os.DirEntry) game {
	replay := decodeReplay(file)
	return getGame(replay)
}

// toon has 4 fields. Three from rep.Details, and one from rep.InitData because the name is unreliable from Details.
func getGame(r *rep.Rep) game {
	Matchup := r.Details.Matchup()
	players := r.Details.Players()
	initData := r.InitData
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
	var you toon
	var opponent toon

	for _, pl := range g.players {
		if _, ok := p.profile[pl.profileID]; ok {
			you = pl
		} else {
			opponent = pl
		}
	}
	fmt.Printf("%12v %6v %6v %-12v  %v %v\n", you.name, you.mmr, opponent.mmr, opponent.name, g.matchup, you.result)
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

func (p *player) updateAllScores(files []os.DirEntry) {
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

func (p *player) updateScore(files []os.DirEntry) (game, error) {
	f, err := getLastModified(files)
	if err != nil {
		return game{}, err
	}
	g := fileToGame(f)
	if !g.isCompetitive {
		return g, errors.New("replay is vs the A.I. (computer)")
	}
	winner := getWinner(g)
	p.setScore(winner.profileID, g.matchup)
	return g, nil
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
	p.ZvP = [2]uint8{}
	p.ZvT = [2]uint8{}
	p.ZvZ = [2]uint8{}
	p.total = [2]uint8{}
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
func sortFilesModTime(files []os.DirEntry) []os.DirEntry {
	sort.Slice(files, func(i, j int) bool {
		info, _ := files[i].Info()
		info2, _ := files[j].Info()
		return info.ModTime().Before(info2.ModTime())
	})
	return files
}

// Sort by modification time in descending order
func sortFilesModTimeDesc(files []os.DirEntry) []os.DirEntry {
	sort.Slice(files, func(i, j int) bool {
		info, _ := files[i].Info()
		info2, _ := files[j].Info()
		return info2.ModTime().Before(info.ModTime())
	})
	return files
}

func getLastModified(files []os.DirEntry) (os.DirEntry, error) {
	if len(files) == 0 {
		return nil, errors.New("no files ending with .SC2Replay found")
	}
	files = sortFilesModTimeDesc(files)
	return files[0], nil
}

func getLastModifiedGame(files []os.DirEntry) game {
	files = sortFilesModTimeDesc(files)

	for _, f := range files {
		g := fileToGame(f)
		if !g.isCompetitive {
			continue
		}
		return g
	}
	return game{}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func redirectError(err error) {
	redirectErr := "Redirecting program to not pull MMR from battlenet API.\n" +
		"MMR will be obtained from your local replay file now."
	fmt.Println(redirectErr)
	fmt.Println()
	log.Println(err)
}
