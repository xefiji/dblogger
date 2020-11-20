logs_all:
	docker-compose -f docker-compose.yml -f examples/docker-compose.yml logs -f --tail=100
stop_all:
	docker-compose -f docker-compose.yml -f examples/docker-compose.yml stop
run_all:
	docker-compose -f docker-compose.yml -f examples/docker-compose.yml up -d
ps_all:
	docker-compose -f docker-compose.yml -f examples/docker-compose.yml ps
build:
	docker-compose build producer
