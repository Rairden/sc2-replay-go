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

func getCredentials() error {
	resp, _ := http.Get(cfg.OAuth2Creds)
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 || resp.ContentLength == 0 {
		err := fmt.Errorf("could not connect to web server\n" +
			"\tStatus Code: %v\n" +
			"\tContent-Length: %v\n\n" +
			"You could register your own Client ID for free at " +
			"https://develop.battle.net/documentation/guides/getting-started\n\n" +
			"Then put the client ID/pass in the cfg.toml file.", resp.StatusCode, resp.ContentLength)
		return err
	}

	type credentials struct {
		ClientID     string `json:"clientID"`
		ClientSecret string `json:"clientSecret"`
	}

	var creds credentials
	json.Unmarshal(body, &creds)

	cfg.clientID = creds.ClientID
	cfg.clientSecret = creds.ClientSecret
	return nil
}

func getBattleNetClient(ID, secret string) *http.Client {
	config := &clientcredentials.Config{
		ClientID:     ID,
		ClientSecret: secret,
		TokenURL:     "https://us.battle.net/oauth/token",
	}

	// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/summary?locale=en_US&access_token=xxx
	client := config.Client(context.Background())
	return client
}

// Set ladderID if not set
func (p *player) setLadderID(client *http.Client) {
	pl := p.profile[cfg.mainToon]
	if pl.ladderID == "" {
		ladderSummaryAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/summary?locale=en_US",
			pl.region, pl.regionID, pl.realmID, pl.profileID)

		apiLadderID := getLadderSummary(client, ladderSummaryAPI, pl.race)
		pl.ladderID = apiLadderID
	}
}

// Returns 0 if data is invalid (nil, 0, -36400), status code != 200, or body is empty.
// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/298683?locale=en_US&access_token=xxx
func (p *player) getMmrAPI(client *http.Client) int64 {
	pl := p.profile[cfg.mainToon]

	if pl.ladderID != "" {
		ladderAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/%s?locale=en_US",
			pl.region, pl.regionID, pl.realmID, pl.profileID, pl.ladderID)

		return getLadder(client, ladderAPI)
	}
	return 0
}

func getLadder(client *http.Client, url string) int64 {
	var lad ladder.Struct

	resp, err := client.Get(url)
	if err != nil {
		return 0
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0
	}

	json.Unmarshal(body, &lad)

	if resp.StatusCode != 200 || len(body) == 0 || len(lad.RanksAndPools) == 0 {
		return 0
	}

	pools := lad.RanksAndPools[0]
	return int64(pools.Mmr)
}

// Returns the ladderID
func getLadderSummary(client *http.Client, url, race string) string {
	var ls laddersummary.Struct

	resp, err := client.Get(url)
	if err != nil {
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
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
				return e.LadderID
			}
		}
	}
	return ""
}
