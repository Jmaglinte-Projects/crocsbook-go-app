### How to run docker for local testing

docker run --rm -p 8080:8080 --env-file .env -e MYSQL_HOST=host.docker.internal crocsbook-api:latest
