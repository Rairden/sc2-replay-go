package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"math"
	"os"
)

func (p *player) setMMR(g game) int64 {
	files, _ := os.ReadDir(cfg.dir)
	MMR := p.getMMR(g)

	if MMR > 0 {
		p.MMR = MMR
		if len(files) == 1 {
			p.startMMR = MMR
		}
		return MMR
	}
	return 0
}

func (p *player) getMMR(g game) int64 {
	for _, pl := range g.players {
		if _, ok := p.profile[pl.name]; ok {
			return pl.mmr
		}
	}
	return 0
}

func (p *player) setStartMMR(files []fs.FileInfo) int64 {
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

func writeData(fullPath string, data string) {
	file, e := os.Create(fullPath)
	check(e)
	defer file.Close()
	file.WriteString(data)
	file.Sync()
}

func (p *player) writeMMRdiff() {
	files, _ := ioutil.ReadDir(cfg.dir)
	if p.startMMR == 0 || numFiles(files) == 1 {
		writeData(mmrDiffTxt, "+0 MMR\n")
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
	writeData(mmrDiffTxt, result)
}

func (p *player) calcMMRdiffAPI(apiMMR int64) {
	writeMMRdiff(p.MMR - apiMMR)
}

func (p *player) writeWinRate() {
	wr := float64(p.getWins()) / float64(p.getTotalGames()) * 100
	winrate := fmt.Sprintf("%.f%%\n", math.Round(wr))
	writeData(winrateTxt, winrate)
}

func (p *player) getTotalGames() uint8 {
	x := p.ZvP[0] + p.ZvP[1]
	y := p.ZvT[0] + p.ZvT[1]
	z := p.ZvZ[0] + p.ZvZ[1]
	return x + y + z
}

func (p *player) getWins() uint8 {
	x := p.ZvP[0]
	y := p.ZvT[0]
	z := p.ZvZ[0]
	return x + y + z
}
