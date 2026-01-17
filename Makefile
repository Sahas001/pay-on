postgres:
	docker run --name postgres18 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:18

createdb:
	docker exec -it postgres18 createdb --username=root --owner=root pay_on
	
dropdb:
	docker exec -it postgres18 dropdb pay_on

migrateup:
	migrate -path internal/database/migration -database "postgresql://root:secret@localhost:5432/pay_on?sslmode=disable" -verbose up

migratedown:
	migrate -path internal/database/migration -database "postgresql://root:secret@localhost:5432/pay_on?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server
