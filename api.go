package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	"sc2-replay-go/api/ladder"
	"sc2-replay-go/api/laddersummary"
)

type BattlenetError struct {
	resp       *http.Response
	body       []byte
	entries    []laddersummary.ShowCaseEntries
	ranksPools []ladder.RanksAndPools
	err        error
}

func (b *BattlenetError) Error() string {
	return fmt.Sprintf("%v\n" +
		"\tStatus Code: %v\n" +
		"\tContent-Length: %v\n" +
		"\tLength of body: %v\n" +
		"\tLength of ShowCaseEntries: %v\n" +
		"\tLength of RanksAndPools: %v\n",
		b.err, b.resp.StatusCode, b.resp.ContentLength, len(b.body), len(b.entries), len(b.ranksPools))
}

func getCredentials() error {
	resp, err := http.Get(cfg.OAuth2Creds)
	if err != nil {
		err := fmt.Errorf("couldn't connect to website. Check the OAuth2Creds URL in cfg.toml.\n\t%v", err)
		return err
	}
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 || resp.ContentLength == 0 {
		err := fmt.Errorf("could not connect to web server\n" +
			"\tStatus Code: %v\n" +
			"\tContent-Length: %v\n\n" +
			"You could register your own Client ID for free at " +
			"https://develop.battle.net/documentation/guides/getting-started\n" +
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

func (p *player) setLadderID(client *http.Client) error {
	pl := p.profile[cfg.mainToon]
	ladderSummaryAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/summary?locale=en_US",
		pl.region, pl.regionID, pl.realmID, pl.profileID)

	apiLadderID, err := getLadderSummary(client, ladderSummaryAPI, pl.race)
	if err != nil {
		return err
	}

	pl.ladderID = apiLadderID
	return nil
}

// getMmrAPI returns 0 if the data is invalid (nil, 0, -36400), status code != 200, or the body is empty.
// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/298683?locale=en_US&access_token=xxx
func (p *player) getMmrAPI(client *http.Client) (int64, error) {
	pl := p.profile[cfg.mainToon]

	if pl.ladderID != "" {
		ladderAPI := fmt.Sprintf("https://%s.api.blizzard.com/sc2/profile/%s/%s/%s/ladder/%s?locale=en_US",
			pl.region, pl.regionID, pl.realmID, pl.profileID, pl.ladderID)

		mmr, err := getLadder(client, ladderAPI)
		if err != nil {
			p.setLadderID(client)
			return getLadder(client, ladderAPI)
		}
		return mmr, nil
	}
	return 0, errors.New("ladderID needs to be set first")
}

func getLadder(client *http.Client, url string) (int64, error) {
	var lad ladder.Struct

	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	json.Unmarshal(body, &lad)

	if resp.StatusCode != 200 || len(body) == 0 || len(lad.RanksAndPools) == 0 {
		return 0, &BattlenetError{resp, body, nil, lad.RanksAndPools, errors.New("error getting ladder")}
	}

	pools := lad.RanksAndPools[0]
	return int64(pools.Mmr), nil
}

// getLadderSummary returns ladderID. See file '/api/laddersummary/json/eu-mamont.json' for a good example.
// mamont plays 1v1 w/ two races (zerg, terran). He has a distinct ladderID for each 1v1 race.
// range over x.showCaseEntries. Look at all the "1v1" and match the race from the cfg.toml file.
func getLadderSummary(client *http.Client, url, race string) (string, error) {
	var ls laddersummary.Struct

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	json.Unmarshal(body, &ls)
	entries := ls.ShowCaseEntries

	statusCode := resp.StatusCode

	if statusCode != 200 || len(body) == 0 || len(entries) == 0 {
		errMsg := fmt.Sprintf("error getting ladder summary. ")

		if statusCode >= 500 {
			errMsg += fmt.Sprintf("The blizzard API is down (%v). ", statusCode)
		}
		return "", &BattlenetError{resp, body, entries, nil, errors.New(errMsg)}
	}

	for _, e := range entries {
		if e.Team.LocalizedGameMode == "1v1" {
			player1 := e.Team.Members[0]
			if player1.FavoriteRace == race {
				return e.LadderID, nil
			}
		}
	}
	return "", errors.New("could not find ladderID for that race. Make sure your cfg.toml file\n" +
		"\thas the correct race for name= and mainToon=")
}
