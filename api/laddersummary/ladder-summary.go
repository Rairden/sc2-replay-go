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
