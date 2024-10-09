package jsonparser

type Registry struct {
	People map[string]string `json:"people"`
}

type MatchEvent struct {
	MatchNumber int    `json:"match_number"`
	Name        string `json:"name"`
}

type Info struct {
	BallsPerOver    int                    `json:"balls_per_over"`
	City            string                 `json:"city"`
	Dates           []string               `json:"dates"`
	Gender          string                 `json:"gender"`
	MatchEvent      MatchEvent             `json:"event,omitempty"`
	MatchType       string                 `json:"match_type"`
	MatchTypeNumber int                    `json:"match_type_number"`
	Officials       map[string]interface{} `json:"officials"`
	Outcome         map[string]interface{} `json:"outcome"`
	Overs           int                    `json:"overs"`
	PlayerOfMatch   []string               `json:"player_of_match"`
	Players         map[string][]string    `json:"players"`
	Registry        Registry               `json:"registry"`
	Season          any                    `json:"season,omitempty"`
	SuperSubs       map[string]string      `json:"supersubs"`
	TeamType        string                 `json:"team_type"`
	Teams           []string               `json:"teams"`
	Toss            map[string]string      `json:"toss"`
	Venue           string                 `json:"venue"`
}

type Run struct {
	Batter int `json:"batter"`
	Extras int `json:"extras"`
	Total  int `json:"total"`
}

type Wicket struct {
	Kind      string `json:"kind"`
	PlayerOut string `json:"player_out"`
	//Fielder   []map[string]string `json:"fielders"`  need to work on it later.:TODO: V2
}

type Delivery struct {
	Batter     string   `json:"batter"`
	Bowler     string   `json:"bowler"`
	NonStriker string   `json:"non_striker"`
	Runs       Run      `json:"runs"`
	Wicket     []Wicket `json:"wickets"`
}

type MatchOver struct {
	OverCount  int        `json:"over"`
	Deliveries []Delivery `json:"deliveries"`
}

type Target struct {
	Over int `json:"overs"`
	Runs int `json:"runs"`
}

type Innings struct {
	Team string      `json:"team"`
	Over []MatchOver `json:"overs"`
	//Target Target      `json:"target"`
}
