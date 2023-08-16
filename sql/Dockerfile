FROM mysql:8.0.33

COPY ./initdb.sql /docker-entrypoint-initdb.d/

ENV MYSQL_DATABASE=Gin_todo
ENV MYSQL_USER=root
ENV MYSQL_PASSWORD=Pastibisa
ENV MYSQL_ROOT_PASSWORD=root

CMD ["mysqld"]

