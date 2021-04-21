package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"log"
	"net/http"
	. "sc2-api/laddersummary"
	"sc2-replay-go/api/ladder"
)

// https://stackoverflow.com/questions/55755929/go-convert-interface-to-map
// https://stackoverflow.com/questions/55382177/type-interface-does-not-support-indexing
func getMMR1() float64 {
	var info map[string]interface{}
	json.Unmarshal(ladder.Ladder2JSON, &info)

	ranksAndPools := info["ranksAndPools"].([]interface{})
	b := ranksAndPools[0].(map[string]interface{})
	mmr := b["mmr"].(float64)

	return mmr
}

func unmarshalLadder() int {
	var ls ladder.Ladder
	io, _ := ioutil.ReadFile("ladder/json/us-Gixxasaurus.json")

	json.Unmarshal(io, &ls)

	if len(ls.RanksAndPools) == 0 {
		return 0
	}
	return ls.RanksAndPools[0].Mmr
}

func unmarshalLadderSummary(race string) (string, string) {
	var ls LadderSummary
	io, _ := ioutil.ReadFile("laddersummary/json/us-Gixxasaurus.json")

	json.Unmarshal(io, &ls)
	entries := ls.ShowCaseEntries

	for _, e := range entries {
		if e.Team.LocalizedGameMode == "1v1" {
			player := e.Team.Members[0]
			if player.FavoriteRace == race {
				return player.Name, e.LadderID
			}
		}
	}

	return "", ""
}

func getLadder(client *http.Client, url string) int {
	var ladder ladder.Ladder

	// Use the client to get some data.
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

func getLadderSummary(client *http.Client, url, race string) string {
	var ls LadderSummary

	// Use the client to get some data and print result.
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
				// player.profile[cfg.main].name = player1.Name
				return e.LadderID
			}
		}
	}
	return ""
}

func oauth2config() {
	// curl -u {client_id}:{client_secret} -d grant_type=client_credentials https://us.battle.net/oauth/token
	// curl -u 632b0e2b3f0a4d64abf4060794fca015:eR5qWtmpyzM4OWzRHqXzhkCwokOq8rEI -d grant_type=client_credentials https://us.battle.net/oauth/token
	// https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/1?locale=en_US&access_token=UStCUHiOt8SMsXS3pueWFLlpZqVBqiaUgv

	// Setup oauth2 config
	conf := &clientcredentials.Config{
		ClientID:     "632b0e2b3f0a4d64abf4060794fca015",
		ClientSecret: "eR5qWtmpyzM4OWzRHqXzhkCwokOq8rEI",
		TokenURL:     "https://us.battle.net/oauth/token",
	}

	// Get a HTTP client
	client := conf.Client(context.Background())

	// Use the client to get some data and print result.
	resp, err := client.Get("https://us.api.blizzard.com/sc2/profile/1/1/1331332/ladder/298683?locale=en_US")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// log.Println("response: ", string(body))

	var info map[string]interface{}
	json.Unmarshal(body, &info)

	// Print the output from info map.
	fmt.Println(info["ranksAndPools"])
}
