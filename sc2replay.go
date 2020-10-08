package main

import (
	"fmt"
	"github.com/icza/s2prot/rep"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

var (
	regex = regexp.MustCompile("^(Gixxasaurus|Rairden)$")
	dir   = "/home/erik/scratch/replays2/"
	// regex2  = regexp.MustCompile("^(1331332|4545534)$")
	// dir2    = "/run/media/erik/storage/SC2/replayBackup/"
	player  = Player{[2]uint8{0, 0}, [2]uint8{0, 0}, [2]uint8{0, 0}}
	matchup = NIL
)

const (
	ZvP = uint8(iota)
	ZvT
	ZvZ
	NIL
)

type Player struct {
	ZvP [2]uint8
	ZvT [2]uint8
	ZvZ [2]uint8
}

func main() {
	files, _ := ioutil.ReadDir(dir)
	fileCnt := numFiles(files)
	player.updateScore(files)

	debugPrint()
	saveAllFiles()

	for {
		time.Sleep(1 * time.Second)

		if fileCnt == numFiles(files) {
			files, _ = ioutil.ReadDir(dir)
			continue
		}

		fileCnt = numFiles(files)
		player.updateScore(files)

		var stringToWrite string
		switch matchup {
		case ZvP:
			stringToWrite = matchupToString(&player.ZvP)
		case ZvT:
			stringToWrite = matchupToString(&player.ZvT)
		case ZvZ:
			stringToWrite = matchupToString(&player.ZvZ)
		}

		saveFile(stringToWrite)
		debugPrint()
	}
}

func (p *Player) resetScores() {
	p.ZvP[0] = 0
	p.ZvP[1] = 0
	p.ZvT[0] = 0
	p.ZvT[1] = 0
	p.ZvZ[0] = 0
	p.ZvZ[1] = 0
}

func (p *Player) updateScore(files []os.FileInfo) {
	p.resetScores()

	for _, file := range files {
		r, err := rep.NewFromFileEvts(dir+file.Name(), false, false, false)

		if err != nil {
			fmt.Printf("Failed to open file: '%v'\n", err)
			return
		}
		defer r.Close()

		matchup := r.Details.Matchup()
		players := r.Details.Players()

		mu := createMatchup(&matchup)

		if players[0].Result().Name == "Victory" {
			player.SetScore(mu, &players[0].Name)
		} else {
			player.SetScore(mu, &players[1].Name)
		}
	}
}

func saveFile(str string) {
	currDir, _ := os.Getwd()
	pwd := currDir + filepath.Join("/")

	var ZvX string
	switch matchup {
	case ZvP:
		ZvX = "ZvP.txt"
	case ZvT:
		ZvX = "ZvT.txt"
	case ZvZ:
		ZvX = "ZvZ.txt"
	}
	fmt.Printf("file = %v\n", ZvX)

	file, e1 := os.Create(pwd + ZvX)
	check(e1)
	defer file.Close()

	_, e2 := file.WriteString(str)
	check(e2)
	file.Sync()
}

func saveAllFiles() {
	currDir, _ := os.Getwd()
	pwd := currDir + filepath.Join("/")

	ZvP_txt, e1 := os.Create(pwd + "ZvP.txt")
	ZvT_txt, e2 := os.Create(pwd + "ZvT.txt")
	ZvZ_txt, e3 := os.Create(pwd + "ZvZ.txt")

	check(e1)
	check(e2)
	check(e3)

	defer ZvP_txt.Close()
	defer ZvT_txt.Close()
	defer ZvZ_txt.Close()

	ZvP_txt.WriteString(matchupToString(&player.ZvP))
	ZvT_txt.WriteString(matchupToString(&player.ZvT))
	ZvZ_txt.WriteString(matchupToString(&player.ZvZ))

	ZvP_txt.Sync()
	ZvT_txt.Sync()
	ZvZ_txt.Sync()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func matchupToString(ZvX *[2]uint8) string {
	win := strconv.Itoa(int(ZvX[0]))
	loss := strconv.Itoa(int(ZvX[1]))
	str := fmt.Sprintf("%2s - %s\n", win, loss)
	return str
}

func (p *Player) SetScore(mu uint8, name *string) {
	switch mu {
	case ZvP:
		incScore(name, &p.ZvP)
	case ZvT:
		incScore(name, &p.ZvT)
	case ZvZ:
		incScore(name, &p.ZvZ)
	}
}

func incScore(name *string, ZvX *[2]uint8) {
	if regex.MatchString(*name) {
		ZvX[0]++
	} else {
		ZvX[1]++
	}
}

func createMatchup(mu *string) uint8 {
	if *mu == "PvZ" || *mu == "ZvP" {
		return ZvP
	}
	if *mu == "TvZ" || *mu == "ZvT" {
		return ZvT
	}
	if *mu == "ZvZ" {
		return ZvZ
	}
	return NIL
}

func numFiles(files []os.FileInfo) int {
	return len(files)
}

func debugPrint() {
	fmt.Println(player.ZvP)
	fmt.Println(player.ZvT)
	fmt.Println(player.ZvZ)
}
