.PHONY: setup
setup: teardown
	@docker-compose -f docker-compose.yaml up --build --wait -d
	@curl http://localhost:5000/info/10

.PHONY: teardown
teardown:
	@docker-compose -f docker-compose.yaml down -v --remove-orphans

.PHONY: test-postgres
test-postgres: setup
	./gonkey -db_dsn "postgresql://testing_user:testing_password@localhost:5432/testing_db?sslmode=disable" -debug -host http://localhost:5000 -tests ./cases/postgres
	make teardown

.PHONY: test-aerospike
test-aerospike: setup
	./gonkey -debug  -fixtures ./fixtures/ -db-type aerospike -aerospike_host "localhost:3000/test" -host http://localhost:5000 -tests ./cases/aerospike
	make teardown