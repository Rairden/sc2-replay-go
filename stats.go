package main

import (
	"fmt"
	"github.com/icza/s2prot/rep"
	"io/ioutil"
	"math"
	"os"
)

func (p *Player) setMMR(r *rep.Rep) int {
	type toon struct {
		Name string
		mmr int64
	}

	var players []toon
	initData := r.InitData.UserInitDatas

	for i := 0; i < 2; i++ {
		p1 := toon{Name: initData[i].Name(), mmr: initData[i].MMR()}
		players = append(players, p1)
	}

	if isMyName(&players[0].Name) {
		mmr := players[0].mmr
		if p.isInvalidMMR(int(mmr)) {
			return 0
		}
		p.MMR = int(mmr)
	} else {
		mmr := players[1].mmr
		if p.isInvalidMMR(int(mmr)) {
			return 0
		}
		p.MMR = int(mmr)
	}

	return p.MMR
}

func (p *Player) isInvalidMMR(mmr int) bool {
	if mmr <= 0 {
		p.MMR = 0
		return true
	}
	return false
}

func writeData(fullPath string, data string) {
	file, e := os.Create(fullPath)
	check(e)
	defer file.Close()
	file.WriteString(data)
	file.Sync()
}

func (p *Player) writeMMRdiff() {
	files, _ := ioutil.ReadDir(cfg.dir)
	if p.startMMR == 0 || numFiles(files) == 1 {
		writeData(MMRdiff_txt, "+0 MMR\n")
		return
	}
	writeMMRdiff(p.startMMR - p.MMR)
}

func writeMMRdiff(diff int) {
	var result string
	if diff <= 0 {
		result = fmt.Sprintf("+%v MMR\n", diff)
	} else {
		result = fmt.Sprintf("-%v MMR\n", diff)
	}
	writeData(MMRdiff_txt, result)
}

func (p *Player) calcMMRdiffAPI(currMMR int) {
	writeMMRdiff(p.MMR - currMMR)
}

func (p *Player) writeWinRate() {
	wr := float64(p.getWins()) / float64(p.getTotalGames()) * 100
	winrate := fmt.Sprintf("%.f%%\n", math.Round(wr))
	writeData(winrate_txt, winrate)
}

func (p *Player) getTotalGames() uint8 {
	x := p.ZvP[0] + p.ZvP[1]
	y := p.ZvT[0] + p.ZvT[1]
	z := p.ZvZ[0] + p.ZvZ[1]
	return x + y + z
}

func (p *Player) getWins() uint8 {
	x := p.ZvP[0]
	y := p.ZvT[0]
	z := p.ZvZ[0]
	return x + y + z
}
