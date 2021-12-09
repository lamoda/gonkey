.PHONY: setup
setup:
	@docker-compose -f docker-compose.yaml up --build -d
	@curl http://localhost:5000/info/10

.PHONY: teardown
teardown:
	@docker-compose -f docker-compose.yaml down -v --remove-orphans

.PHONY: test
test: setup
	gonkey -db_dsn "postgresql://testing_user:testing_password@localhost:5432/testing_db?sslmode=disable" -debug -host http://localhost:5000 -tests ./cases/
	make teardown
