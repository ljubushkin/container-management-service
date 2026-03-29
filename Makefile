DB_DSN=postgres://postgres:postgres@localhost:5432/containers?sslmode=disable

goose-up:
	goose -dir migrations postgres "$(DB_DSN)" up

goose-down:
	goose -dir migrations postgres "$(DB_DSN)" down

goose-status:
	goose -dir migrations postgres "$(DB_DSN)" status

goose-create:
	goose -dir migrations create $(name) sql