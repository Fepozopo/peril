#!/bin/bash

case "$1" in
    start)
        echo "Starting RabbitMQ container..."
        docker run -d --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management
        ;;
    stop)
        echo "Stopping RabbitMQ container..."
        docker stop rabbitmq
        ;;
    logs)
        echo "Fetching logs for RabbitMQ container..."
        docker logs -f rabbitmq
        ;;
    start-stomp)
        echo "Starting RabbitMQ STOMP container..."
        docker run -d --rm --name rabbitmq-stomp -p 5672:5672 -p 15672:15672 -p 61613:61613 rabbitmq-stomp
        ;;
    stop-stomp)
        echo "Stopping RabbitMQ STOMP container..."
        docker stop rabbitmq-stomp
        ;;
    logs-stomp)
        echo "Fetching logs for RabbitMQ STOMP container..."
        docker logs -f rabbitmq-stomp
        ;;
    *)
        echo "Usage: $0 {start|stop|logs}"
        exit 1
esac
