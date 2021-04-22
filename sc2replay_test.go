package main

import "testing"

func Test_createMatchup(t *testing.T) {
	tests := []struct {
		input string
		want  uint8
	}{
		{"PvZ", ZvP},
		{"TvZ", ZvT},
		{"ZvP", ZvP},
		{"ZvT", ZvT},
		{"ZvZ", ZvZ},
	}

	for _, test := range tests {
		if matchup := getMatchup(test.input); matchup != test.want {
			t.Errorf("want: %v\n", test.want)
			t.Errorf(" got: %v\n\n", matchup)
		}
	}
}

func Test_matchupToString(t *testing.T) {
	type player struct {
		ZvX [2]uint8
	}
	tests := []struct {
		player player
		want   string
	}{
		{player{[2]uint8{0, 0}}, " 0 - 0\n"},
		{player{[2]uint8{5, 3}}, " 5 - 3\n"},
		{player{[2]uint8{10, 4}}, "10 - 4\n"},
		{player{[2]uint8{8, 12}}, " 8 - 12\n"},
		{player{[2]uint8{11, 13}}, "11 - 13\n"},
	}

	for _, test := range tests {
		if got := scoreToString(&test.player.ZvX); got != test.want {
			t.Errorf("want: %v", test.want)
			t.Errorf(" got: %v\n", got)
		}
	}
}

func BenchmarkMain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		main()
	}
}
