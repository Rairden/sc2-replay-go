package ladder

type Ladder struct {
	LadderTeams             []LadderTeams           `json:"ladderTeams"`
	AllLadderMemberships    []AllLadderMemberships  `json:"allLadderMemberships"`
	LocalizedDivision       string                  `json:"localizedDivision"`
	League                  string                  `json:"league"`
	CurrentLadderMembership CurrentLadderMembership `json:"currentLadderMembership"`
	RanksAndPools           []RanksAndPools         `json:"ranksAndPools"`
}
type TeamMembers struct {
	ID           string `json:"id"`
	Realm        int    `json:"realm"`
	Region       int    `json:"region"`
	DisplayName  string `json:"displayName"`
	FavoriteRace string `json:"favoriteRace"`
}
type LadderTeams struct {
	TeamMembers   []TeamMembers `json:"teamMembers"`
	PreviousRank  int           `json:"previousRank"`
	Points        int           `json:"points"`
	Wins          int           `json:"wins"`
	Losses        int           `json:"losses"`
	Mmr           int           `json:"mmr"`
	JoinTimestamp int           `json:"joinTimestamp"`
}
type AllLadderMemberships struct {
	LadderID          string `json:"ladderId"`
	LocalizedGameMode string `json:"localizedGameMode"`
	Rank              int    `json:"rank"`
}
type CurrentLadderMembership struct {
	LadderID          string `json:"ladderId"`
	LocalizedGameMode string `json:"localizedGameMode"`
}
type RanksAndPools struct {
	Rank      int `json:"rank"`
	Mmr       int `json:"mmr"`
	BonusPool int `json:"bonusPool"`
}

var Ladder3JSON = []byte(`{
  "ladderTeams": [],
  "allLadderMemberships": [],
  "ranksAndPools": []
}`)

var Ladder2JSON = []byte(`{
    "ladderTeams": [],
    "allLadderMemberships": [
        {
            "ladderId": "298982",
            "localizedGameMode": "1v1 Platinum",
            "rank": 40
        }
    ],
    "ranksAndPools": []
}`)
