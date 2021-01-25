FROM proxysql/proxysql:latest
# Install mysql client for getting proxysql stats
RUN apt-get update && apt-get install -y mysql-client

COPY proxysql.cnf /etc/proxysql.cnf