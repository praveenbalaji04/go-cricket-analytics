

Migration commands

./migrate -database postgres://localhost:5432/z_cricket?sslmode=disable -path ../cricket/internal/db/migrations up

./migrate create -ext sql -dir ../cricket/internal/db/migrations -seq add_player_source_id

./migrate -database postgres://localhost:5432/z_cricket?sslmode=disable -path ../cricket/internal/db/migrations force 3