package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"cricket/cmd/app"
	jsonparser "cricket/pkg/json_parser"
)

type baseStruct struct {
	Meta    json.RawMessage      `json:"meta"`
	Info    jsonparser.Info      `json:"info"`
	Innings []jsonparser.Innings `json:"innings"`
}

type eventSql struct {
	FileId        int
	MatchId       int
	Name          string
	Date          string
	TeamA         int
	TeamB         int
	PlayingXiAIds []int
	PlayingXiBIds []int
	Venue         string
	Toss          string
	Overs         int
	MatchType     string
}

type wicketSql struct {
	player int
	kind   string
	bowler int
	event  int
}

type ballInfoSql struct {
	Event       int
	Over        int
	Ball        int
	BattingTeam int
	Batsman     int
	Bowler      int
	NonStriker  int
	StrikerRun  int
	ExtraRun    int
	Wickets     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func ReadData(service *app.App) {
	directoryName := "t20s_male_json"
	files, err := os.ReadDir(directoryName)
	if err != nil {
		panic("error in reading json directory")
	}

	skippedItems := 0
	parsedMatchesBrokenCount := 0
	for _, e := range files {
		jsonFilePath := fmt.Sprintf("%s/%s", directoryName, e.Name())
		content, err := os.ReadFile(jsonFilePath)
		if strings.Contains(jsonFilePath, "README.txt") {
			// ignore readme file
			continue
		}
		breakParsing := false

		if err != nil {
			service.Logger.Info(
				"error in fetching data",
				zap.String("file path", jsonFilePath),
				zap.Error(err),
			)
			//continue
			panic("err")
		}

		var jsonData baseStruct
		err = json.Unmarshal(content, &jsonData)
		if err != nil {
			service.Logger.Info(
				"error in unmarshalling json data",
				zap.Error(err),
				zap.String("file", jsonFilePath),
			)
			skippedItems += 1
			continue
		}

		exists := isMatchDataExists(jsonData.Info.MatchTypeNumber, service)
		if exists {
			service.Logger.Info("skipping file as it is already exists", zap.String("filename", jsonFilePath))
			continue
		}
		service.Logger.Info("successfully unmarshalled file", zap.String("file", jsonFilePath))

		service.Logger.Info("calling save team function", zap.Any("teams", jsonData.Info.Teams))
		teamInfo, err := saveTeam(jsonData.Info.Teams, service.DB)
		if err != nil {
			service.Logger.Info(
				"error in saving team",
				zap.Error(err),
				zap.Any("teams", jsonData.Info.Teams),
				zap.Int("match_id", jsonData.Info.MatchTypeNumber))
			panic("error in saving team!!!!")
		}

		service.Logger.Info(
			"get or create team",
			zap.Int("match_id", jsonData.Info.MatchTypeNumber),
			zap.Any("teams", jsonData.Info.Teams),
			zap.Any("saved or read team id", teamInfo),
		)

		teamPlayersId := make(map[int][]int)
		for teamName, players := range jsonData.Info.Players {
			playersMapping := make(map[string]string)

			sourceIds := make([]string, 0)
			for _, player := range players {
				if strings.Contains(player, "(2)") || strings.Contains(player, "(3)") {
					breakParsing = true
					break
				}
				sourceId := jsonData.Info.Registry.People[player]
				playersMapping[sourceId] = player
				sourceIds = append(sourceIds, sourceId)
			}
			if breakParsing {
				break
			}
			teamId := teamInfo[teamName]
			err := savePlayersBulk(playersMapping, teamId, service.DB)
			if err != nil {
				log.Printf("error in saving players: %s", err.Error())
				return
			}

			playerIds, err := getPlayersBasedOnSourceId(sourceIds, service.DB)
			if err != nil {
				service.Logger.Info(
					"error in getting players based on source id",
					zap.Error(err),
					zap.String("team name", teamName),
				)
				panic("error in getting players from source id")
			}
			teamPlayersId[teamId] = playerIds
		}
		if breakParsing {
			// skip the entire match
			parsedMatchesBrokenCount += 1
			continue
		}

		filename := strings.Split(filepath.Base(jsonFilePath), ".json")[0]
		basePath, _ := strconv.Atoi(filename)

		tossAsString := fmt.Sprintf(
			"%v won the toss and chose to %v", jsonData.Info.Toss["winner"], jsonData.Info.Toss["decision"])

		eventData := eventSql{
			FileId:        basePath,
			MatchId:       jsonData.Info.MatchTypeNumber,
			Name:          jsonData.Info.MatchEvent.Name,
			Date:          jsonData.Info.Dates[0],
			TeamA:         teamInfo[jsonData.Info.Teams[0]],
			TeamB:         teamInfo[jsonData.Info.Teams[1]],
			PlayingXiAIds: teamPlayersId[teamInfo[jsonData.Info.Teams[0]]],
			PlayingXiBIds: teamPlayersId[teamInfo[jsonData.Info.Teams[1]]],
			Venue:         jsonData.Info.Venue,
			Toss:          tossAsString, // adding it as a string for now.
			Overs:         jsonData.Info.Overs,
			MatchType:     jsonData.Info.MatchType,
		}

		eventId, err := saveEvent(eventData, service.DB)
		if err != nil {
			service.Logger.Info(
				"error in storing event",
				zap.Int("match id", jsonData.Info.MatchTypeNumber),
				zap.String("event", eventData.Name),
				zap.Error(err))
			panic("error in storing event")
			// cannot continue without event ID
		}
		service.Logger.Info(
			"completed saving event for match",
			zap.String("event", eventData.Name),
			zap.Int("match id", jsonData.Info.MatchTypeNumber),
		)

		teamPlayers := getPlayersBasedOnMatch(jsonData.Info.MatchTypeNumber, service)
		// plan how to skip if already inserted
		for _, data := range jsonData.Innings {
			teamId := teamInfo[data.Team]
			ballCount := 0
			for _, overInfo := range data.Over {
				overCount := overInfo.OverCount
				for i, deliveryInfo := range overInfo.Deliveries {
					var wicketId int
					if deliveryInfo.Wicket != nil {
						wicketData := wicketSql{
							player: teamPlayers[deliveryInfo.Wicket[0].PlayerOut],
							bowler: teamPlayers[deliveryInfo.Bowler],
							kind:   deliveryInfo.Wicket[0].Kind,
							event:  eventId,
						}

						wicketId, err = addWicket(wicketData, service)
						if err != nil {
							service.Logger.Info(
								"error in saving wicket",
								zap.Error(err),
								zap.String("player out", deliveryInfo.Wicket[0].PlayerOut),
								zap.String("bowler", deliveryInfo.Bowler),
								zap.Any("other detail", teamPlayers),
								zap.Int("match id", jsonData.Info.MatchTypeNumber),
							)
							panic("error in saving wickets")
						}
					}
					ballInfo := ballInfoSql{
						Event:       eventId,
						Over:        overCount,
						Ball:        i,
						BattingTeam: teamId,
						Batsman:     teamPlayers[deliveryInfo.Batter],
						Bowler:      teamPlayers[deliveryInfo.Bowler],
						NonStriker:  teamPlayers[deliveryInfo.NonStriker],
						StrikerRun:  deliveryInfo.Runs.Batter,
						ExtraRun:    deliveryInfo.Runs.Extras,
						Wickets:     wicketId,
					}
					ballCount += 1
					err = saveBallInfo(ballInfo, service)
					if err != nil {
						service.Logger.Info("error in saving ball", zap.Error(err))
						panic("error in saving ball")
					}
					service.Logger.Info("saved ball info successfully")
				}
			}
			service.Logger.Info(
				"saved all the information of the match",
				zap.Int("match id", jsonData.Info.MatchTypeNumber),
				zap.Int("team id", teamId),
				zap.Int("over count", len(data.Over)),
				zap.Int("ballCount", ballCount),
			)
		}
	}
	fmt.Println("skipped items", skippedItems, "parsedMatchesBrokenCount", parsedMatchesBrokenCount)
}

func isMatchDataExists(matchId int, service *app.App) bool {
	sqlQuery := `SELECT EXISTS(SELECT 1 FROM event where match_id = $1)`
	var exists bool
	err := service.DB.QueryRow(context.TODO(), sqlQuery, matchId).Scan(&exists)
	if err != nil {
		service.Logger.Info("error in checking match data exists", zap.Int("match id", matchId))
		return false
	}
	return exists
}

func addWicket(wicketSqlData wicketSql, service *app.App) (int, error) {
	sqlQuery := `INSERT INTO wicket (player, bowler, event, kind) VALUES (@player, @bowler, @event, @kind) RETURNING id`
	namedArgs := pgx.NamedArgs{
		"player": wicketSqlData.player,
		"bowler": wicketSqlData.bowler,
		"event":  wicketSqlData.event,
		"kind":   wicketSqlData.kind,
	}

	var id int
	err := service.DB.QueryRow(context.Background(), sqlQuery, namedArgs).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func getPlayersBasedOnMatch(matchId int, service *app.App) map[string]int {
	sqlQuery := `SELECT p.id, p.name FROM event AS e JOIN player as p ON p.id = ANY(e.playing_11_a_ids) or p.id = ANY(e.playing_11_b_ids) where match_id = (@match_id)`

	namedArgs := pgx.NamedArgs{"match_id": matchId}
	rows, err := service.DB.Query(context.Background(), sqlQuery, namedArgs)
	if err != nil {
		service.Logger.Info(
			"error in fetching players based on match",
			zap.Error(err),
			zap.Int("match_id", matchId),
		)
		return nil
	}
	players := make(map[string]int)
	if rows != nil {
		for rows.Next() {
			var id int
			var name string
			err = rows.Scan(&id, &name)
			if err != nil {
				service.Logger.Info("players fetched, but error in looping rows")
			}
			players[name] = id
		}
	}
	defer rows.Close()
	return players
}

func saveTeam(teamNames []string, dbInstance *pgxpool.Pool) (map[string]int, error) {
	response := make(map[string]int)
	for _, name := range teamNames {
		sqlQuery := `INSERT INTO team (name) VALUES (@name) ON CONFLICT (name) DO NOTHING RETURNING (id)`
		namedArgs := pgx.NamedArgs{"name": name}
		var id int
		err := dbInstance.QueryRow(context.TODO(), sqlQuery, namedArgs).Scan(&id)
		if err != nil {
			sqlQuery = `SELECT id from team WHERE name = (@name)`
			err = dbInstance.QueryRow(context.TODO(), sqlQuery, namedArgs).Scan(&id)
			if err != nil {
				return nil, err
			}
		}
		response[name] = id
	}
	return response, nil
}

func savePlayersBulk(playersInfo map[string]string, teamId int, dbInstance *pgxpool.Pool) error {
	// No need to return response, this is just storing all the players
	sqlQuery := `INSERT INTO player (name, source_id, team_id) VALUES (@name, @source_id, @team_id) ON CONFLICT (source_id) DO NOTHING RETURNING (id)`

	batch := pgx.Batch{}
	for sourceId, player := range playersInfo {
		namedArgs := pgx.NamedArgs{"name": player, "source_id": sourceId, "team_id": teamId}
		batch.Queue(sqlQuery, namedArgs)
	}

	results := dbInstance.SendBatch(context.Background(), &batch)

	defer func(results pgx.BatchResults) {
		_ = results.Close()
	}(results)

	for player := range playersInfo {
		_, err := results.Exec()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				log.Printf("player %s already exists, skipping this player", player)
				continue
			}
			log.Printf("Unable to save player %v", err)
		}
	}
	return nil
}

func getPlayersBasedOnSourceId(sourceId []string, dbInstance *pgxpool.Pool) ([]int, error) {
	sqlQuery := `SELECT id FROM player as p WHERE p.source_id = ANY($1)`
	// pass the value directly in query when using $. use namedArgs only when using name in sqlQuery
	rows, err := dbInstance.Query(context.TODO(), sqlQuery, sourceId)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var IDs []int
	for rows.Next() {
		var id int
		_ = rows.Scan(&id)
		IDs = append(IDs, id)
	}
	return IDs, nil
}

func saveEvent(event eventSql, dbInstance *pgxpool.Pool) (int, error) {
	sqlQuery := `
		INSERT INTO event (
			file_id,
			match_id,
			name,
			date,
			team_a,
			team_b,
			playing_11_a_ids,
			playing_11_b_ids,
			venue,
			toss,
			overs,
			match_type
		)
		VALUES (
			@file_id,
			@match_id,
			@name,
			@date,
			@team_a,
			@team_b,
			@playing_11_a_ids,
			@playing_11_b_ids,
			@venue,
			@toss,
			@overs,
			@match_type
		)
		ON CONFLICT (match_id) DO UPDATE SET name = @name RETURNING (id)`
	/*
		using DO UPDATE ON CONFLICT is not a correct approach, modify it later as per below link
		https://dba.stackexchange.com/questions/129522/how-to-get-the-id-of-the-conflicting-row-in-upsert?newreg=73012b692b4f484d8406e4f67dd98ea6
	*/

	namedArgs := pgx.NamedArgs{
		"file_id":          event.FileId,
		"match_id":         event.MatchId,
		"name":             event.Name,
		"date":             event.Date,
		"team_a":           event.TeamA,
		"team_b":           event.TeamB,
		"playing_11_a_ids": event.PlayingXiAIds,
		"playing_11_b_ids": event.PlayingXiBIds,
		"venue":            event.Venue,
		"toss":             event.Toss,
		"overs":            event.Overs,
		"match_type":       event.MatchType,
	}

	var id int
	err := dbInstance.QueryRow(context.TODO(), sqlQuery, namedArgs).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func saveBallInfo(ballInfo ballInfoSql, service *app.App) error {
	sqlQuery := `
		INSERT INTO ball_info (
							event,
							over,
							ball,
							batting_team,
							batsman,
							bowler,
							non_striker,
							striker_run,
							extra_run,
							wicket
							)
							VALUES (
							@event,
							@over,
							@ball,
							@batting_team,
							@batsman,
							@bowler,
							@non_striker,
							@striker_run,
							@extra_run,
							@wicket
		        )`

	namedArgs := pgx.NamedArgs{
		"event":        ballInfo.Event,
		"over":         ballInfo.Over,
		"ball":         ballInfo.Ball,
		"batting_team": ballInfo.BattingTeam,
		"batsman":      ballInfo.Batsman,
		"bowler":       ballInfo.Bowler,
		"non_striker":  ballInfo.NonStriker,
		"striker_run":  ballInfo.StrikerRun,
		"extra_run":    ballInfo.ExtraRun,
		//"wicket":       nil,
	}
	if ballInfo.Wickets != 0 {
		namedArgs["wicket"] = ballInfo.Wickets
	}

	rows, err := service.DB.Query(context.TODO(), sqlQuery, namedArgs)
	defer rows.Close()
	if err != nil {
		service.Logger.Info("error in saving ball info", zap.Error(err))
		return err
	}
	return nil
}
