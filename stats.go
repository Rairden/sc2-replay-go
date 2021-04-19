package main

import (
	"fmt"
	"github.com/icza/s2prot/rep"
	"io/ioutil"
	"math"
	"os"
)

func (p *Player) setMMR(r *rep.Rep) float64 {
	players := r.Details.Players()
	metadata := r.Metadata.Players()

	if isPlayer(&players[0].Name) {
		mmr := metadata[0].Value("MMR")
		if p.isInvalidMMR(mmr) {
			return 0
		}
		p.MMR = mmr.(float64)
	} else {
		mmr := metadata[1].Value("MMR")
		if p.isInvalidMMR(mmr) {
			return 0
		}
		p.MMR = mmr.(float64)
	}

	return p.MMR
}

func (p *Player) isInvalidMMR(mmr interface{}) bool {
	if mmr == nil || mmr.(float64) < 0 {
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

	var result string
	diff := p.startMMR - p.MMR

	if diff <= 0 {
		diff = math.Abs(diff)
		result = fmt.Sprintf("+%.f MMR\n", diff)
	} else {
		result = fmt.Sprintf("-%.f MMR\n", diff)
	}
	writeData(MMRdiff_txt, result)
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
