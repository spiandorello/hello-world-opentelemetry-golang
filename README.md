```bash
docker run -d --name hello-world-trace-db -p 3306:3306 \
    -e MYSQL_ROOT_PASSWORD=mysqlpwd mysql
```

```bash
docker exec -i hello-world-trace-db mysql -uroot -pmysqlpwd < ./database.sql
```

```bash
docker run -d --name jaeger \
    -p 6831:6831/udp \
    -p 16686:16686 \
    -p 14268:14268 \
    jaegertracing/all-in-one
```