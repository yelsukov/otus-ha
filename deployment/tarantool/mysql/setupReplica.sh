#!/bin/bash
printf "\n================================================================\n"
printf "=                             MySQl                            =\n"
printf "================================================================\n"
printf "[%s] Configuring MySQL master node...\n" "$(date)"
# Creates user and gives him the rights to be the replica reader and data dumper
mysql --user="$MASTER_USER" --password="$MASTER_PASSWORD" --host=mysqldb -P3306 \
  -e "CREATE USER '$REPL_USER'@'%' IDENTIFIED BY '$REPL_PASSWORD';" \
  -e "GRANT REPLICATION CLIENT, REPLICATION SLAVE, RELOAD, PROCESS ON *.* TO '$REPL_USER'@'%';
  GRANT SELECT ON \`${DB_NAME}\`.* TO '$REPL_USER'@'%';" \
  -e "FLUSH PRIVILEGES;"

printf "\n\n================================================================\n"
printf "=                        Master status                         =\n"
printf "================================================================\n"
mysql --user="$MASTER_USER" --password="$MASTER_PASSWORD" --host="master_db" -A \
  -e "SHOW MASTER STATUS\G"