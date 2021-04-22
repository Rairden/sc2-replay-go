package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"math"
	"os"
)

func (p *Player) setMMR(g Game) int64 {
	files, _ := os.ReadDir(cfg.dir)

	if isMyName(g.players[0].name) {
		MMR := g.players[0].mmr
		if p.isInvalidMMR(MMR) {
			p.MMR = 0
			return 0
		}
		p.MMR = MMR
		if len(files) == 1 {
			p.startMMR = MMR
		}
	} else {
		MMR := g.players[1].mmr
		if p.isInvalidMMR(MMR) {
			p.MMR = 0
			return 0
		}
		p.MMR = MMR
		if len(files) == 1 {
			p.startMMR = MMR
		}
	}

	return p.MMR
}

func (p *Player) setStartMMR(files []fs.FileInfo) int64 {
	files = sortFilesModTime(files)

	for _, file := range files {
		game := fileToGame(file)
		p.startMMR = p.setMMR(game)
		if p.startMMR <= 0 {
			continue
		} else {
			return p.MMR
		}
	}

	return 0
}

func (p *Player) isInvalidMMR(mmr int64) bool {
	if mmr <= 0 {
		p.startMMR, p.MMR = 0, 0
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

func writeMMRdiff(diff int64) {
	var result string
	if diff <= 0 {
		diff *= -1
		result = fmt.Sprintf("+%d MMR\n", diff)
	} else {
		result = fmt.Sprintf("-%d MMR\n", diff)
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
