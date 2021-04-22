package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	. "sc2-api/laddersummary"
	"sc2-replay-go/api/ladder"
)

func getLadder(client *http.Client, url string) int {
	var ladder ladder.Ladder

	resp, err := client.Get(url)
	if err != nil {
		// log.Fatal(err)
		return 0
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// log.Fatal(err)
		return 0
	}

	json.Unmarshal(body, &ladder)

	if resp.StatusCode != 200 || len(body) == 0 || len(ladder.RanksAndPools) == 0 {
		return 0
	}

	pools := ladder.RanksAndPools[0]
	return pools.Mmr
}

// returns the ladderId
func getLadderSummary(client *http.Client, url, race string) string {
	var ls LadderSummary

	resp, err := client.Get(url)
	if err != nil {
		// log.Fatal(err)
		return ""
	}


	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println()
	if err != nil {
		// log.Fatal(err)
		return ""
	}

	json.Unmarshal(body, &ls)
	entries := ls.ShowCaseEntries

	if resp.StatusCode != 200 || len(body) == 0 || len(entries) == 0 {
		return ""
	}

	for _, e := range entries {
		if e.Team.LocalizedGameMode == "1v1" {
			player1 := e.Team.Members[0]
			if player1.FavoriteRace == race {
				// Consider taking name from here as Details has garbage: github.com/icza/sc2prot/rep.Details  --> "&lt;QOSQO&gt;FoXx"
				// players.profile[cfg.main].name = player1.Name
				return e.LadderID
			}
		}
	}
	return ""
}
