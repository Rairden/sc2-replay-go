package laddersummary

type LadderSummary struct {
	ShowCaseEntries      []ShowCaseEntries      `json:"showCaseEntries"`
	PlacementMatches     []interface{}          `json:"placementMatches"`
	AllLadderMemberships []AllLadderMemberships `json:"allLadderMemberships"`
}
type Members struct {
	FavoriteRace string `json:"favoriteRace"`
	Name         string `json:"name"`
	PlayerID     string `json:"playerId"`
	Region       int    `json:"region"`
}
type Team struct {
	LocalizedGameMode string    `json:"localizedGameMode"`
	Members           []Members `json:"members"`
}
type ShowCaseEntries struct {
	LadderID              string `json:"ladderId"`
	Team                  Team   `json:"team"`
	LeagueName            string `json:"leagueName"`
	LocalizedDivisionName string `json:"localizedDivisionName"`
	Rank                  int    `json:"rank"`
	Wins                  int    `json:"wins"`
	Losses                int    `json:"losses"`
}
type AllLadderMemberships struct {
	LadderID          string `json:"ladderId"`
	LocalizedGameMode string `json:"localizedGameMode"`
	Rank              int    `json:"rank"`
}

var LadderSummaryJSON = []byte(`{
  "showCaseEntries": [
    {
      "ladderId": "298683",
      "team": {
        "localizedGameMode": "1v1",
        "members": [
          {
            "favoriteRace": "zerg",
            "name": "Gixxasaurus",
            "playerId": "1331332",
            "region": 1
          }
        ]
      },
      "leagueName": "DIAMOND",
      "localizedDivisionName": "Ramsey Victor",
      "rank": 24,
      "wins": 27,
      "losses": 25
    }
  ],
  "placementMatches": [],
  "allLadderMemberships": [
    {
      "ladderId": "298683",
      "localizedGameMode": "1v1 Diamond",
      "rank": 24
    }
  ]
}`)

var LadderSummaryJSON2 = []byte(`{
  "showCaseEntries": [
    {
      "ladderId": "237200",
      "team": {
        "localizedGameMode": "1v1",
        "members": [
          {
            "favoriteRace": "zerg",
            "name": "mamont",
            "playerId": "8904344",
            "region": 2
          }
        ]
      },
      "leagueName": "DIAMOND",
      "localizedDivisionName": "Judicator Tau",
      "rank": 1,
      "wins": 79,
      "losses": 60
    },
    {
      "ladderId": "237154",
      "team": {
        "localizedGameMode": "1v1",
        "members": [
          {
            "favoriteRace": "terran",
            "name": "mamont",
            "playerId": "8904344",
            "region": 2
          }
        ]
      },
      "leagueName": "PLATINUM",
      "localizedDivisionName": "Colossus Uncle",
      "rank": 79,
      "wins": 0,
      "losses": 1
    }
  ],
  "placementMatches": [],
  "allLadderMemberships": [
    {
      "ladderId": "237200",
      "localizedGameMode": "1v1 Diamond",
      "rank": 1
    },
    {
      "ladderId": "237154",
      "localizedGameMode": "1v1 Platinum",
      "rank": 79
    }
  ]
}`)

var LadderSummaryJSON3 = []byte(`{
  "showCaseEntries": [],
  "placementMatches": [],
  "allLadderMemberships": []
}`)
