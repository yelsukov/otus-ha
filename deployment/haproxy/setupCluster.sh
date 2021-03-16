#!/bin/bash

CREATE_SERVICE_USERS="CREATE USER '$REPL_USER'@'%' IDENTIFIED BY '$REPL_PASSWORD';
  GRANT REPLICATION SLAVE ON *.* TO '$REPL_USER'@'%';
  CREATE USER 'haproxy'@'%';
  FLUSH PRIVILEGES;"

INIT_SLAVE="CHANGE MASTER TO MASTER_HOST='master_db',MASTER_USER='$REPL_USER',MASTER_PASSWORD='$REPL_PASSWORD',MASTER_AUTO_POSITION=1;
  START SLAVE;"

INIT_SLAVE_SEMI="INSTALL PLUGIN rpl_semi_sync_slave SONAME 'semisync_slave.so';
  SET GLOBAL rpl_semi_sync_slave_enabled=1; STOP SLAVE IO_THREAD; START SLAVE IO_THREAD;"

printf "\n================================================================\n"
printf "=                             MySQl                            =\n"
printf "================================================================\n"
printf "[%s] Configuring MySQL master node...\n" "$(date)"
mysql --user="$MASTER_USER" --password="$MASTER_PASSWORD" --host=master_db -P3306 \
  -e "$CREATE_SERVICE_USERS" \
  -e "INSTALL PLUGIN rpl_semi_sync_master SONAME 'semisync_master.so';
  SET GLOBAL rpl_semi_sync_master_enabled=1; SET GLOBAL rpl_semi_sync_master_wait_for_slave_count=3; SET GLOBAL rpl_semi_sync_master_timeout=5000;"

for HOST in "slave1_db" "slave2_db" "slave3_db"
do
  printf "[%s] Configuring MySQL %s node...\n" "$(date)" "${HOST}"
  mysql --user="$MASTER_USER" --password="$MASTER_PASSWORD" --host="${HOST}" -P3306 \
    -e "$INIT_SLAVE" -e "$CREATE_SERVICE_USERS" -e "$INIT_SLAVE_SEMI"
done
sleep 1

printf "\n\n================================================================\n"
printf "=                        Master status                         =\n"
printf "================================================================\n"
mysql --user="$MASTER_USER" --password="$MASTER_PASSWORD" --host="master_db" -A -e "SHOW MASTER STATUS\G"
printf "\nAdditional Settings:\n"
mysql --user="$MASTER_USER" --password="$MASTER_PASSWORD" --host="master_db" -A \
 -e "SHOW VARIABLES LIKE 'rpl_semi_sync%';" \
 -e "SHOW VARIABLES LIKE 'binlog_format';" \
 -e "SHOW VARIABLES LIKE '%gtid%';"


for HOST in "slave1_db" "slave2_db" "slave3_db"
do
	printf "\n================================================================\n"
  printf "=                  Slave status of %s                   =\n" "${HOST}"
  printf "================================================================\n"
  mysql --user="$MASTER_USER" --password="$MASTER_PASSWORD" --host="${HOST}" -A  -e "SHOW SLAVE STATUS\G"
  printf "\nAdditional Settings:\n"
  mysql --user="$MASTER_USER" --password="$MASTER_PASSWORD" --host="${HOST}" -A \
   -e "SHOW VARIABLES LIKE 'rpl_semi_sync%';" \
   -e "SHOW VARIABLES LIKE 'binlog_format';" \
   -e "SHOW VARIABLES LIKE '%gtid%';"
done