package main

import (
	"fmt"
	"math"
	"os"
)

// getReplayMMR sets the player MMR and returns the value. And returns 0 if you haven't
// played in a long time, it's your first 2-3 placement matches, or an A.I. replay.
func (p *player) getReplayMMR(g game) int64 {
	var MMR int64

	if !g.isCompetitive {
		return 0
	}

	for _, pl := range g.players {
		if _, ok := p.profile[pl.profileID]; ok {
			MMR = pl.mmr
		}
	}

	if MMR > 0 {
		p.MMR = MMR
		return MMR
	}

	p.MMR = 0
	return 0
}

// setStartMMR returns 0 if the data is invalid (nil, 0, -36400) or vs the A.I.
func (p *player) setStartMMR(files []os.DirEntry) int64 {
	files = sortFilesModTime(files)

	for _, file := range files {
		game := fileToGame(file)
		MMR := p.getReplayMMR(game)

		if MMR <= 0 || !game.isCompetitive {
			continue
		}
		p.startMMR = MMR
		return MMR
	}
	p.startMMR = 0
	return 0
}

func writeData(fullPath, data string) {
	file, e := os.Create(fullPath)
	check(e)
	defer file.Close()
	file.WriteString(data)
	file.Sync()
}

func (p *player) writeMMRdiff(start, end int64, fileCnt ...int) {
	if !cfg.useAPI {
		if p.startMMR == 0 || fileCnt[0] == 1 {
			writeData(mmrDiffTxt, "+0 MMR\n")
			return
		}
	}

	if p.startMMR > 0 && end > 0 {
		diff := start - end
		var result string
		if diff <= 0 {
			diff *= -1
			result = fmt.Sprintf("+%d MMR\n", diff)
		} else {
			result = fmt.Sprintf("-%d MMR\n", diff)
		}
		writeData(mmrDiffTxt, result)
		return
	}

	writeData(mmrDiffTxt, "+0 MMR\n")
}

func (p *player) writeWinRate() {
	if p.getTotalGames() == 0 {
		writeData(winrateTxt, "0%\n")
		return
	}
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
	x := p.xvp[0]
	y := p.xvt[0]
	z := p.xvz[0]
	return x + y + z
}

func (p *player) getLosses() uint8 {
	x := p.xvp[1]
	y := p.xvt[1]
	z := p.xvz[1]
	return x + y + z
}
