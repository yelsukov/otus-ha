#!/bin/bash

# Replica user added to master and slaves to rebuild cluster on failover
CREATE_SERVICE_USERS="CREATE USER '$REPL_USER'@'%' IDENTIFIED BY '$REPL_PASSWORD';
  GRANT REPLICATION SLAVE ON *.* TO '$REPL_USER'@'%';
  CREATE USER '$MONITOR_USER'@'%' IDENTIFIED BY '$MONITOR_PASSWORD';
  GRANT REPLICATION CLIENT ON *.* TO '$MONITOR_USER'@'%';
  CREATE USER '$PROXY_USER'@'%' IDENTIFIED BY '$PROXY_PASSWORD';
  GRANT INSERT, SELECT, UPDATE, DELETE, LOCK TABLES, EXECUTE, CREATE, ALTER, INDEX, REFERENCES ON otus_ha.* TO '$PROXY_USER'@'%';
  CREATE USER '${ORC_USER}'@'%' IDENTIFIED BY '$ORC_PASSWORD';
  GRANT SUPER, PROCESS, REPLICATION SLAVE, RELOAD ON *.* TO '${ORC_USER}'@'%';
  GRANT SELECT ON mysql.slave_master_info TO '${ORC_USER}'@'%';
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
sleep 1

for HOST in "slave1_db" "slave2_db" "slave3_db"
do
  printf "[%s] Configuring MySQL %s node...\n" "$(date)" "${HOST}"
  mysql --user="$SLAVE_USER" --password="$SLAVE_PASSWORD" --host="${HOST}" -P3306 \
    -e "$INIT_SLAVE" -e "$CREATE_SERVICE_USERS" -e "$INIT_SLAVE_SEMI"
done
sleep 1

printf "\n\n================================================================\n"
printf "=                           ProxySQL                           =\n"
printf "================================================================\n"
printf "[%s] Waiting for ProxySQL service...\n" "$(date)"
RC=1
while [ $RC -eq 1 ]
do
    if ! mysqladmin ping -hproxysql -P6032 -ufirestarter -pstartfire  > /dev/null 2>&1;
    then
      sleep 1
    else
        RC=0
        echo "ProxySQL Server is alive!"
    fi
done
printf "[%s] Configuring ProxySQL...\n" "$(date)"
mysql --user=firestarter --password=startfire -hproxysql -P6032 \
  -e "INSERT INTO mysql_users (username,password,active,default_hostgroup) values ('$PROXY_USER','$PROXY_PASSWORD',1,1);
  LOAD MYSQL USERS TO RUNTIME;
  SAVE MYSQL USERS TO DISK;
  UPDATE global_variables SET variable_value='${MONITOR_USER}' WHERE variable_name='mysql-monitor_username';
  UPDATE global_variables SET variable_value='${MONITOR_PASSWORD}' WHERE variable_name='mysql-monitor_password';
  LOAD MYSQL VARIABLES TO RUNTIME;
  SAVE MYSQL VARIABLES TO DISK;
  UPDATE global_variables SET variable_value='${PROXY_ADMIN}:${PROXY_ADMIN_PASSWORD}' WHERE variable_name='admin-admin_credentials';
  LOAD ADMIN VARIABLES TO RUNTIME;
  SAVE ADMIN VARIABLES TO DISK;
  INSERT INTO mysql_servers (hostgroup_id,hostname,port,max_replication_lag) VALUES (1,'master_db',3306,1), (2,'slave1_db',3306,1), (2,'slave2_db',3306,1), (2,'slave3_db',3306,1);
  LOAD MYSQL SERVERS TO RUNTIME;
  SAVE MYSQL SERVERS TO DISK;"
printf "[%s] ProxySQL Provisioning has been completed" "$(date)"


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
  mysql --user="$SLAVE_USER" --password="$SLAVE_PASSWORD" --host="${HOST}" -A  -e "SHOW SLAVE STATUS\G"
  printf "\nAdditional Settings:\n"
  mysql --user="$SLAVE_USER" --password="$SLAVE_PASSWORD" --host="${HOST}" -A \
   -e "SHOW VARIABLES LIKE 'rpl_semi_sync%';" \
   -e "SHOW VARIABLES LIKE 'binlog_format';" \
   -e "SHOW VARIABLES LIKE '%gtid%';"
done