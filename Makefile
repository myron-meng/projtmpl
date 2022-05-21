run:
	@go mod download
	@go mod tidy
	@cd cmd/api && go run ./main.go

build:
	@go mod download
	@go mod tidy
	@go mod vendor
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o projtmpl cmd/api/main.go

fmt:
	@gofmt -s -l -w .
	@goimports -w -local projtmpl/ .

migrate:
	@migrate -database "mysql://${MIGRATE_DB_USERNAME}:${MIGRATE_DB_PASSWORD}@tcp(${MIGRATE_DB_HOST})/${DB_NAME}" -path "internal/database/migration" up

sqlgen:
	@sqlboiler --add-soft-deletes -p model -c internal/database/sqlboiler.toml mysql
