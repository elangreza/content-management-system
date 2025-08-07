#!make
include .env
	
run:
	go run cmd/server/main.go

FILENAME?=file-name

migrate:
	@read -p  "up or down or version? " MODE; \
	migrate -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOSTNAME}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSL}" -path ${MIGRATION_FOLDER} $$MODE
	
migrate-create:
	@read -p  "What is the name of migration? " NAME; \
	migrate create -ext sql -tz Asia/Jakarta -dir ${MIGRATION_FOLDER} -format "20060102150405" $$NAME

gen:
	go generate ./...

.PHONY: migrate migrate-create run