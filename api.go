package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	. "sc2-api/laddersummary"
	"sc2-replay-go/api/ladder"
)

func getBattleNetClient() *http.Client {
	config := &clientcredentials.Config{
		ClientID:     cfg.apiClientId,
		ClientSecret: cfg.apiClientPass,
		TokenURL:     "https://us.battle.net/oauth/token",
	}

	// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/summary?locale=en_US&access_token=xxx
	client := config.Client(context.Background())
	return client
}

// set ladderId if not set
func setLadderId(client *http.Client) {
	p := player.profile[cfg.mainToon]
	if player.profile[cfg.mainToon].ladderId == "" {
		ladderSummaryAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/summary?locale=en_US",
			p.region, p.regionId, p.realmId, p.profileId)

		apiLadderId := getLadderSummary(client, ladderSummaryAPI, p.race)
		p.ladderId = apiLadderId
	}
}

// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/298683?locale=en_US&access_token=xxx
func getMMR(client *http.Client) int {
	p := player.profile[cfg.mainToon]
	ladderAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/%s?locale=en_US",
		p.region, p.regionId, p.realmId, p.profileId, p.ladderId)

	return getLadder(client, ladderAPI)
}

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
				// players.profile[cfg.mainToon].name = player1.Name
				return e.LadderID
			}
		}
	}
	return ""
}
