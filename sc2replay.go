package main

import (
	"errors"
	"flag"
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
	xvp, xvt, xvz, total [2]uint8
	startMMR, MMR        int64
	profile              map[string]*profile
}

type player2 struct {
	*player
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
	return p.getReplayMMR(lastGame), nil
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
	race      *rep.Race
	mmr       int64
	result    string
}

func main() {
	player := setup(cfgToml)
	player2 := &player2{player, nil}

	replays := flag.String("print", cfg.dir, "Prints a summary of your replays.")
	flag.Parse()

	myFlag := flag.Lookup("print")

	if isFlagPassed("print") {
		if myFlag.Value.String() != "" {
			player.printAllGames(*replays)
		} else {
			player.printAllGames(myFlag.DefValue)
		}
		os.Exit(0)
	}

	fmt.Printf("\nChecking the directory '%v' \nevery %v ms for new SC2 replays...\n\n", cfg.dir, cfg.updateTime)

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
			mainNoAPI(pl.player)
		}
	}

	client := getBattleNetClient(cfg.clientID, cfg.clientSecret)
	pl.client = client
	_, err := pl.setLadderID(client)
	if err != nil {
		redirectError(err)
		mainNoAPI(pl.player)
	}

	// Allow user to start with a non-empty replay folder
	if len(files) > 0 {
		pl.setStartMMR(files)
		pl.updateAllScores(files)
		pl.writeWinRate()
	} else {
		saveResetStats()
	}

	mmr, err := pl.getMmrAPI(client)
	pl.writeMMRdiff(pl.startMMR, mmr)
	if err != nil {
		redirectError(err)
		mainNoAPI(pl.player)
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
	cfg.useAPI = false

	// Allow user to start with a non-empty replay folder
	if len(files) > 0 {
		pl.setStartMMR(files)
		pl.updateAllScores(files)
		newestGame := getLastModifiedGame(files)
		mmr := pl.getReplayMMR(newestGame)
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
			if len(files) <= fileCnt {
				fileCnt = len(files)
			}
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
		p.saveFile(p.getOpponent(game).race.Name)

		if p.startMMR == 0 {
			p.setStartMMR(files)
		}

		mmr, err := usr.getMMR(files)
		if err != nil {
			if _, ok := err.(*BattlenetError); ok {
				redirectError(err)
				mainNoAPI(p)
			}
			if err == ErrPromoted {
				newestFile, _ := getLastModified(files)
				lastGame := fileToGame(newestFile)
				mmr = p.getReplayMMR(lastGame)
			}
		}

		if cfg.useAPI && err != ErrPromoted {
			if p.MMR == mmr {
				p.retryBattlenet(usr)
			} else {
				p.MMR = mmr
			}
		}

		p.writeMMRdiff(p.startMMR, p.MMR, len(files))
		p.printGame(game)
	}
}

func (p *player) retryBattlenet(usr user) {
	var mmr int64
	var delay int64 = 3000

	// try 3x, each separated by 3, 6, 12 seconds.
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		mmr, _ = usr.getMMR()
		if p.MMR == mmr {
			delay *= 2
		} else {
			p.MMR = mmr
			return
		}
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

func decodeReplay(file os.DirEntry, dir ...string) *rep.Rep {
	if len(dir) > 0 {
		cfg.dir = dir[0]
	}
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
	matchup := r.Details.Matchup()
	players := r.Details.Players()
	initData := r.InitData
	userInitDatas := initData.UserInitDatas

	// Only InitData shows it's an A.I. (computer) match at 'x.InitData.Struct.gameDescription.gameOptions.competitive'
	isCompetitive := initData.GameDescription.GameOptions.CompetitiveOrRanked()

	var toons []toon

	for i := 0; i < 2; i++ {
		p1 := toon{
			strconv.FormatInt(players[i].Toon.ID(), 10),
			userInitDatas[i].Name(),
			players[i].Race(),
			userInitDatas[i].MMR(),
			players[i].Result().String(),
		}
		toons = append(toons, p1)
	}

	return game{toons, matchup, isCompetitive}
}

func getWinner(g game) toon {
	if g.players[0].result == "Victory" {
		return g.players[0]
	}
	return g.players[1]
}

func (p *player) getOpponent(g game) toon {
	toon := g.players[0]
	if !p.isMyID(toon.profileID) {
		return toon
	}
	return g.players[1]
}

func (p *player) printGame(g game) {
	var you toon
	var opponent toon

	if _, ok := p.profile[g.players[0].profileID]; ok {
		you = g.players[0]
		opponent = g.players[1]
	} else {
		you = g.players[1]
		opponent = g.players[0]
	}

	matchup := g.matchup
	if you.profileID != "" {
		matchup = fixMatchup(g.matchup, you.race.Letter)
	}

	fmt.Printf("%12v %6v %6v   %-12v  %v  %v\n",
		you.name, you.mmr, opponent.mmr, opponent.name, matchup, you.result)
}

func (p *player) printAllGames(dir string) {
	reps := getAllReplays(dir)

	for _, r := range reps {
		replay := decodeReplay(r, dir)
		g := getGame(replay)
		p.printGame(g)
	}
}

// fixMatchup returns a reversed string for the players' perspective.
func fixMatchup(mu string, yourRace rune) string {
	if rune(mu[0]) == yourRace {
		return mu
	}
	matchup := []byte(mu)
	sort.Slice(matchup, func(i, j int) bool {
		return true
	})
	return string(matchup)
}

func isTieOrServerCrashed(g game) bool {
	return g.players[0].result == "Tie" || g.players[0].result == "Unknown"
}

func (p *player) updateAllScores(files []os.DirEntry) {
	for _, file := range files {
		g := fileToGame(file)
		if !g.isCompetitive || isTieOrServerCrashed(g) {
			continue
		}
		winner := getWinner(g)
		p.setScore(winner.profileID, p.getOpponent(g).race.Name)
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

	if isTieOrServerCrashed(g) {
		return game{}, errors.New("the game was a tie or the bnet server crashed")
	}

	winner := getWinner(g)
	p.setScore(winner.profileID, p.getOpponent(g).race.Name)
	return g, nil
}

// setScore The ID must be the winner.
func (p *player) setScore(ID, opponentRace string) {
	switch opponentRace {
	case "Protoss":
		p.incScore(ID, &p.xvp)
	case "Terran":
		p.incScore(ID, &p.xvt)
	case "Zerg":
		p.incScore(ID, &p.xvz)
	}
}

func (p *player) incScore(ID string, XvX *[2]uint8) {
	isYou := p.isMyID(ID)

	if isYou {
		XvX[0]++
	} else {
		XvX[1]++
	}
}

func (p *player) resetPlayer() {
	p.xvp = [2]uint8{}
	p.xvt = [2]uint8{}
	p.xvz = [2]uint8{}
	p.total = [2]uint8{}
	p.MMR, p.startMMR = 0, 0
}

func (p *player) isMyID(ID string) bool {
	if _, ok := p.profile[ID]; ok {
		return true
	}
	return false
}

func (p *player) saveFile(opponentRace string) {
	switch opponentRace {
	case "Protoss":
		writeFile(xvpTxt, &p.xvp)
	case "Terran":
		writeFile(xvtTxt, &p.xvt)
	case "Zerg":
		writeFile(xvzTxt, &p.xvz)
	}
}

func (p *player) saveAllFiles() {
	writeFile(xvpTxt, &p.xvp)
	writeFile(xvtTxt, &p.xvt)
	writeFile(xvzTxt, &p.xvz)
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

func scoreToString(XvX *[2]uint8) string {
	win := strconv.Itoa(int(XvX[0]))
	loss := strconv.Itoa(int(XvX[1]))
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
	errorAPI := ""
	get, ok := err.(*BattlenetError)
	if ok {
		errorAPI = fmt.Sprintf("The blizzard API is down (%v). ", get.resp.StatusCode)
	}

	redirectErr := fmt.Sprintf("%sMMR will be obtained from your local replay file now.\n", errorAPI)
	fmt.Println(redirectErr)
	log.Println(err)
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
