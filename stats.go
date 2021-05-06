package main

import (
	"fmt"
	"io/fs"
	"math"
	"os"
)

func (p *player) getReplayMMR(g game) int64 {
	for _, pl := range g.players {
		if _, ok := p.profile[pl.profileID]; ok {
			return pl.mmr
		}
	}
	return 0
}

func (p *player) setMMR(g game) int64 {
	files, _ := os.ReadDir(cfg.dir)
	MMR := p.getReplayMMR(g)

	if MMR > 0 {
		p.MMR = MMR
		if len(files) == 1 {
			p.startMMR = MMR
		}
		return MMR
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
		}
		return p.MMR
	}

	return 0
}

func writeData(fullPath, data string) {
	file, e := os.Create(fullPath)
	check(e)
	defer file.Close()
	file.WriteString(data)
	file.Sync()
}

func (p *player) writeMMRdiff(diff int64) {
	files, _ := os.ReadDir(cfg.dir)

	if !cfg.useAPI {
		if p.startMMR == 0 || numFiles(files) == 1 {
			writeData(mmrDiffTxt, "+0 MMR\n")
			return
		}
	}

	var result string
	if diff <= 0 {
		diff *= -1
		result = fmt.Sprintf("+%d MMR\n", diff)
	} else {
		result = fmt.Sprintf("-%d MMR\n", diff)
	}
	writeData(mmrDiffTxt, result)
}

func (p *player) writeWinRate() {
	wr := float64(p.getWins()) / float64(p.getTotalGames()) * 100
	winrate := fmt.Sprintf("%.f%%\n", math.Round(wr))
	writeData(winrateTxt, winrate)
}

func (p *player) writeTotalWinLoss() {
	writeFile(totalWinLossTxt, &p.total)
}

func (p *player) setTotalWinLoss() {
	p.total[0] = p.getWins()
	p.total[1] = p.getLosses()
}

func (p *player) getTotalGames() uint8 {
	return p.getWins() + p.getLosses()
}

func (p *player) getWins() uint8 {
	x := p.ZvP[0]
	y := p.ZvT[0]
	z := p.ZvZ[0]
	return x + y + z
}

func (p *player) getLosses() uint8 {
	x := p.ZvP[1]
	y := p.ZvT[1]
	z := p.ZvZ[1]
	return x + y + z
}
