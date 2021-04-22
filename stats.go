package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
)

func (p *Player) setMMR(game Game) int64 {
	if isMyName(game.players[0].name) {
		mmr := game.players[0].mmr
		if p.isInvalidMMR(mmr) {
			return 0
		}
		p.MMR = mmr
	} else {
		mmr := game.players[1].mmr
		if p.isInvalidMMR(mmr) {
			return 0
		}
		p.MMR = mmr
	}

	return p.MMR
}

func (p *Player) isInvalidMMR(mmr int64) bool {
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

// FIXME: Solve edge case if their first 1-3 games are invalid mmr (0 or -36400).
// 	Need to set their start MMR somehow.
func (p *Player) writeMMRdiff() {
	files, _ := ioutil.ReadDir(cfg.dir)
	if p.startMMR == 0 || numFiles(files) == 1 {
		writeData(MMRdiff_txt, "+0 MMR\n")
		return
	}
	writeMMRdiff(p.startMMR - p.MMR)
}

func writeMMRdiff(diff int64) {
	var result string
	if diff <= 0 {
		result = fmt.Sprintf("+%v MMR\n", diff)
	} else {
		result = fmt.Sprintf("-%v MMR\n", diff)
	}
	writeData(MMRdiff_txt, result)
}

func (p *Player) calcMMRdiffAPI(apiMMR int64) {
	writeMMRdiff(p.MMR - apiMMR)
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
