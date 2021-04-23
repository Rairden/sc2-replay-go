package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	"sc2-replay-go/api/ladder"
	"sc2-replay-go/api/laddersummary"
)

func getBattleNetClient() *http.Client {
	config := &clientcredentials.Config{
		ClientID:     cfg.apiClientID,
		ClientSecret: cfg.apiClientPass,
		TokenURL:     "https://us.battle.net/oauth/token",
	}

	// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/summary?locale=en_US&access_token=xxx
	client := config.Client(context.Background())
	return client
}

// set ladderID if not set
func (p *player) setLadderID(client *http.Client) {
	pl := p.profile[cfg.mainToon]
	if pl.ladderID == "" {
		ladderSummaryAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/summary?locale=en_US",
			pl.region, pl.regionID, pl.realmID, pl.profileID)

		apiLadderID := getLadderSummary(client, ladderSummaryAPI, pl.race)
		pl.ladderID = apiLadderID
	}
}

// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/298683?locale=en_US&access_token=xxx
func (p *player) getMmrAPI(client *http.Client) int {
	pl := p.profile[cfg.mainToon]
	ladderAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/%s?locale=en_US",
		pl.region, pl.regionID, pl.realmID, pl.profileID, pl.ladderID)

	return getLadder(client, ladderAPI)
}

func getLadder(client *http.Client, url string) int {
	var lad ladder.Struct

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

	json.Unmarshal(body, &lad)

	if resp.StatusCode != 200 || len(body) == 0 || len(lad.RanksAndPools) == 0 {
		return 0
	}

	pools := lad.RanksAndPools[0]
	return pools.Mmr
}

// returns the ladderID
func getLadderSummary(client *http.Client, url, race string) string {
	var ls laddersummary.Struct

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
				// players.profile[cfg.mainToon].name = player1.Name
				return e.LadderID
			}
		}
	}
	return ""
}
