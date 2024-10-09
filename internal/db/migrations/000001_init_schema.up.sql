CREATE TABLE team (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE player (
    id serial PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    team_id INT,
    CONSTRAINT fk_team FOREIGN KEY(team_id) REFERENCES team(id) ON DELETE CASCADE,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE event (
    id serial PRIMARY KEY,
    file_id INT NOT NULL,
    match_id INT NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    date date NOT NULL,
    team_a INT, CONSTRAINT fk_team_a FOREIGN KEY (team_a) REFERENCES team(id) ON DELETE CASCADE,
    team_b INT, CONSTRAINT fk_team_b FOREIGN KEY (team_b) REFERENCES team(id) ON DELETE CASCADE,
    playing_11_a_ids INTEGER[],
    playing_11_b_ids INTEGER[],
    venue VARCHAR(200),
    toss VARCHAR(255) NOT NULL,
    overs INT NOT NULL,
    match_type VARCHAR(100) NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE wicket (
    id serial PRIMARY KEY,
    player INT, CONSTRAINT fk_player FOREIGN KEY (player) REFERENCES player(id) ON DELETE CASCADE,
    kind varchar(100) not null
);

CREATE TABLE ball_info (
    id serial primary key,
    event int, CONSTRAINT fk_event FOREIGN KEY (event) REFERENCES event(id) ON DELETE CASCADE,
    over int not null,
    ball int not null,
    batting_team int, CONSTRAINT fk_batting_team FOREIGN KEY (batting_team) REFERENCES team(id) ON DELETE SET NULL,
    batsman int, CONSTRAINT fk_batsman FOREIGN KEY (batsman) REFERENCES player(id) ON DELETE SET NULL,
    bowler int, CONSTRAINT fk_bowler FOREIGN KEY (bowler) REFERENCES player(id) ON DELETE SET NULL,
    non_striker int, CONSTRAINT fk_non_striker FOREIGN KEY (non_striker) REFERENCES player(id) ON DELETE SET NULL,
    run int not null,
    wicket int, CONSTRAINT fk_wicket FOREIGN KEY (wicket) REFERENCES wicket(id) ON DELETE SET NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE end_result (
    event int, CONSTRAINT fk_event FOREIGN KEY (event) REFERENCES event(id) ON DELETE CASCADE,
    result varchar(500) not null,
    team_won int, CONSTRAINT fk_team_won FOREIGN KEY (team_won) REFERENCES team(id) ON DELETE CASCADE,
    team_a_score int not null,
    team_b_score int not null,
    player_of_the_match int, CONSTRAINT fk_player_of_the_match FOREIGN KEY (player_of_the_match) REFERENCES player(id) ON DELETE SET NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);
