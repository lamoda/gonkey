.PHONY: setup
setup: teardown
	@docker-compose -f docker-compose.yaml up --build -d
	@sleep 2
	@curl http://localhost:5000/info/10

.PHONY: teardown
teardown:
	@docker-compose -f docker-compose.yaml down -v --remove-orphans

.PHONY: test-postgres
test-postgres: setup
	./gonkey -db_dsn "postgresql://testing_user:testing_password@localhost:5432/testing_db?sslmode=disable" -debug -host http://localhost:5000 -tests ./cases/postgres
