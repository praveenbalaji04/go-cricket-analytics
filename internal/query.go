package internal

import (
	"context"
	"fmt"
	"time"

	"cricket/cmd/app"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type pgDB struct {
	db *pgxpool.Pool
}

type WicketResponse struct {
	Caught          int `db:"caught" json:"caught"`
	Bowled          int `db:"bowled" json:"bowled"`
	Stumped         int `db:"stumped" json:"stumped"`
	RunOut          int `db:"run_out" json:"run_out"`
	LBW             int `db:"lbw" json:"lbw"`
	RetiredOut      int `db:"retired_out" json:"retired_out"`
	CaughtAndBowled int `db:"caught_and_bowled" json:"caught_and_bowled"`
}

type TournamentStatsResponse struct {
	Matches      int            `json:"matches"`
	TeamsCount   int            `json:"teams_count"`
	PlayersCount int            `json:"players_count"`
	Boundaries   int            `json:"boundaries"`
	Sixes        int            `json:"sixes"`
	Wickets      WicketResponse `json:"wickets"`
	BallsBowled  int            `json:"balls_bowled"`
	//drawMatches  int
}

func (service pgDB) QueryDB(sqlQuery string) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := service.db.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (service pgDB) QueryCount(sqlQuery string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var count int
	err := service.db.QueryRow(ctx, sqlQuery).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func queryUniqueEvents(dBPool *pgxpool.Pool) ([]string, error) {
	eventQuery := `SELECT DISTINCT(match_type) FROM event`

	rows, err := dBPool.Query(context.TODO(), eventQuery)
	if err != nil {
		return nil, err
	}
	var matchTypes []string

	for rows.Next() {
		var matchType string
		err = rows.Scan(&matchType)
		if err != nil {
			return nil, err
		}
		matchTypes = append(matchTypes, matchType)
	}
	rows.Close()
	return matchTypes, nil
}

func totalMatches(matchType string, dbPool *pgxpool.Pool) (int, error) {
	sqlQuery := `SELECT COUNT(*) FROM event where match_type = @match_type`
	namedArgs := pgx.NamedArgs{"match_type": matchType}
	var matchCount int
	err := dbPool.QueryRow(context.TODO(), sqlQuery, namedArgs).Scan(&matchCount)
	if err != nil {
		return 0, err
	}
	return matchCount, nil
}

func tournamentTeamsCount(matchType string, dbPool *pgxpool.Pool) (int, error) {
	sqlQuery := `
			SELECT COUNT(DISTINCT team) as unique_teams FROM (
			    SELECT event.team_a AS team from event where event.match_type = @match_type
			UNION 
				SELECT event.team_b from event where event.match_type = @match_type
			)as merged_teams;
			`
	namedArgs := pgx.NamedArgs{"match_type": matchType}
	var teamCount int
	err := dbPool.QueryRow(context.TODO(), sqlQuery, namedArgs).Scan(&teamCount)
	if err != nil {
		return 0, err
	}
	return teamCount, nil
}

// returns teams, players distinct count
func tournamentPlayersCount(matchType string, dbPool *pgxpool.Pool) (int, error) {
	sqlQuery := `
			SELECT COUNT(DISTINCT players) as unique_players
			FROM LATERAL (
		            SELECT UNNEST(event.playing_11_a_ids) AS players from event where event.match_type = @match_type
		        UNION 
		            SELECT UNNEST(event.playing_11_b_ids) AS players from event where event.match_type = @match_type)
		    as merged_players;
			`
	namedArgs := pgx.NamedArgs{"match_type": matchType}
	var playersCount int
	err := dbPool.QueryRow(context.TODO(), sqlQuery, namedArgs).Scan(&playersCount)
	if err != nil {
		return 0, err
	}
	return playersCount, nil
}

func tournamentBoundariesCount(matchType string, dbPool *pgxpool.Pool) (int, int, error) {
	sqlQuery := `
				SELECT 
				    COUNT(CASE WHEN bi.striker_run = 4 THEN 1 END) AS boundaries,
				    COUNT(CASE WHEN bi.striker_run = 6 THEN 1 END) AS sixes
				FROM ball_info as bi JOIN event as e on bi.event = e.id and e.match_type = @match_type`
	namedArgs := pgx.NamedArgs{"match_type": matchType}

	var boundariesCount, sixesCount int
	err := dbPool.QueryRow(context.TODO(), sqlQuery, namedArgs).Scan(&boundariesCount, &sixesCount)
	if err != nil {
		return 0, 0, err
	}
	return boundariesCount, sixesCount, nil
}

func tournamentBallsBowled(matchType string, dbPool *pgxpool.Pool) (int, error) {
	sqlQuery := `SELECT COUNT(*) FROM ball_info as bi JOIN event as e ON bi.event = e.id AND e.match_type = @match_type`
	namedArgs := pgx.NamedArgs{"match_type": matchType}

	var ballsBowled int
	err := dbPool.QueryRow(context.TODO(), sqlQuery, namedArgs).Scan(&ballsBowled)
	if err != nil {
		return 0, err
	}
	return ballsBowled, nil
}

func tournamentWickets(matchType string, dbPool *pgxpool.Pool) (WicketResponse, error) {
	sqlQuery := `SELECT 
    				COUNT(CASE WHEN w.kind = 'caught' THEN 1 END) as caught,
    				COUNT(CASE WHEN w.kind = 'bowled' THEN 1 END) as bowled,
    				COUNT(CASE WHEN w.kind = 'stumped' THEN 1 END) as stumped,
    				COUNT(CASE WHEN w.kind = 'run out' THEN 1 END) as run_out,
    				COUNT(CASE WHEN w.kind = 'lbw' THEN 1 END) as lbw,
    				COUNT(CASE WHEN w.kind in ('retired hurt', 'retired out') THEN 1 END) as retired_out,
    				COUNT(CASE WHEN w.kind = 'caught and bowled' THEN 1 END)
				FROM wicket as w JOIN event as e ON w.event = e.id AND e.match_type = @match_type`
	namedArgs := pgx.NamedArgs{"match_type": matchType}

	rows, err := dbPool.Query(context.TODO(), sqlQuery, namedArgs)
	if err != nil {
		return WicketResponse{}, err
	}
	wicketInfo, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[WicketResponse])
	if err != nil {
		fmt.Println(wicketInfo, "wicket info here")
		return WicketResponse{}, err
	}
	return wicketInfo, nil
}

func QueryTournamentStats(appInstance *app.App) (TournamentStatsResponse, error) {
	dbInstance := pgDB{db: appInstance.DB}
	matchTypes, err := queryUniqueEvents(dbInstance.db)
	if err != nil {
		return TournamentStatsResponse{}, err
	}
	stats := TournamentStatsResponse{}

	for _, matchType := range matchTypes {
		appInstance.Logger.Info("fetching total matches")
		matchCount, err := totalMatches(matchType, dbInstance.db)
		if err != nil {
			appInstance.Logger.Info("error in fetching total match", zap.Error(err))
		}
		appInstance.Logger.Info("fetching tournament teams count")
		teamsCount, err := tournamentTeamsCount(matchType, dbInstance.db)
		if err != nil {
			appInstance.Logger.Info("error in fetching total teams count", zap.Error(err))
		}
		appInstance.Logger.Info("fetching players count")
		playersCount, err := tournamentPlayersCount(matchType, dbInstance.db)
		if err != nil {
			appInstance.Logger.Info("error in fetching total players count", zap.Error(err))
		}
		appInstance.Logger.Info("fetching boundary information")
		boundaries, sixes, err := tournamentBoundariesCount(matchType, dbInstance.db)
		if err != nil {
			appInstance.Logger.Info("error in fetching boundary information", zap.Error(err))
		}
		appInstance.Logger.Info("fetching wickets info")
		wicketsInfo, err := tournamentWickets(matchType, dbInstance.db)
		if err != nil {
			appInstance.Logger.Info("error in fetching wickets info", zap.Error(err))
		}
		appInstance.Logger.Info("fetching balls bowled")
		ballsBowled, err := tournamentBallsBowled(matchType, dbInstance.db)
		if err != nil {
			appInstance.Logger.Info("error in fetching balls bowled", zap.Error(err))
		}

		stats.Matches = matchCount
		stats.TeamsCount = teamsCount
		stats.PlayersCount = playersCount
		stats.Boundaries = boundaries
		stats.Sixes = sixes
		stats.BallsBowled = ballsBowled
		stats.Wickets = wicketsInfo
	}
	return stats, err
}
