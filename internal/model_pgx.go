package internal

import (
	"time"
)

type Team struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Player struct {
	ID        int64
	Name      string
	Team      Team
	SourceId  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Toss struct {
	Decision string
	Winner   string
}

type Event struct {
	ID         int64
	FileId     int
	MatchId    int
	Name       string
	Date       time.Time
	TeamA      Team
	TeamB      Team
	PlayingXIA []Player
	PlayingXIB []Player
	Venue      string
	Toss       Toss
	Overs      int
	MatchType  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

//type Run struct {
//	ByBatsman int
//	Extras    int
//	Total     int
//}

type Wicket struct {
	Player Player
	Kind   string
	Bowler Player
	Event  Event
}

type BallInfo struct {
	ID          int64
	Event       Event
	Over        int
	Ball        int
	BattingTeam Team
	Batsman     Player
	Bowler      Player
	NonStriker  Player
	StrikerRun  int
	ExtraRun    int
	Wickets     Wicket
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EndResult match result cannot be calculated every time by calculating ball info.
// once match is completed, calculate basic details and store it in below struct
type EndResult struct {
	Event            Event
	Result           string
	TeamWon          Team
	TeamAScore       int
	TeamBScore       int
	PlayerOfTheMatch Player
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
