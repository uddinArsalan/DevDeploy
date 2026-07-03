rabbitmq:
	docker run -it --rm --hostname rabbitmq-devdeploy --name rabbitmq -p 5672:5672 -p 15672:15672 -v rabbitmq-data:/var/lib/rabbitmq rabbitmq:4-management