#!/bin/bash

printf "\n================================================================\n"
printf "=                             MySQl                            =\n"
printf "================================================================\n"
printf "[%s] Configuring MySQL master node...\n" "$(date)"
mysql --user="$MASTER_USER" --password="$MASTER_PASSWORD" --host=master_db -P3306 \
  -e "CREATE USER '$REPL_USER'@'%' IDENTIFIED BY '$REPL_PASSWORD';
  GRANT REPLICATION SLAVE ON *.* TO '$REPL_USER'@'%';" \
  -e "CREATE USER '$MONITOR_USER'@'%' IDENTIFIED BY '$MONITOR_PASSWORD';
  GRANT REPLICATION CLIENT ON *.* TO '$MONITOR_USER'@'%';" \
  -e "CREATE USER '$PROXY_USER'@'%' IDENTIFIED BY '$PROXY_PASSWORD';
  GRANT INSERT, SELECT, UPDATE, DELETE, LOCK TABLES, EXECUTE, CREATE, ALTER, INDEX, REFERENCES ON otus_ha.* TO '$PROXY_USER'@'%';
  FLUSH PRIVILEGES;"


printf "[%s] Configuring MySQL slave node...\n" "$(date)"
mysql --user="$SLAVE_USER" --password="$SLAVE_PASSWORD" --host=slave_db -P3306 \
  -e "CHANGE MASTER TO MASTER_HOST='master_db',MASTER_USER='$REPL_USER',MASTER_PASSWORD='$REPL_PASSWORD';
  START SLAVE;" \
  -e "CREATE USER '$MONITOR_USER'@'%' IDENTIFIED BY '$MONITOR_PASSWORD';
  GRANT REPLICATION CLIENT ON *.* TO '$MONITOR_USER'@'%';" \
  -e "CREATE USER '$PROXY_USER'@'%' IDENTIFIED BY '$PROXY_PASSWORD';
  GRANT INSERT, SELECT, UPDATE, DELETE, LOCK TABLES, EXECUTE, CREATE, ALTER, INDEX, REFERENCES ON otus_ha.* TO '$PROXY_USER'@'%';
  FLUSH PRIVILEGES;"
printf "[%s] MySQL Provisioning has been completed!\n" "$(date)"

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
  INSERT INTO mysql_servers (hostgroup_id,hostname,port,max_replication_lag) VALUES (1,'master_db',3306,1), (2,'slave_db',3306,1);
  LOAD MYSQL SERVERS TO RUNTIME;
  SAVE MYSQL SERVERS TO DISK;"
printf "[%s] ProxySQL Provisioning has been completed" "$(date)"


printf "\n\n================================================================\n"
printf "=                        Master status                         =\n"
printf "================================================================\n"
mysql --user="$MASTER_USER" --password="$MASTER_PASSWORD" --host="master_db" -A \
  -e "SHOW MASTER STATUS\G"

printf "\n================================================================\n"
printf "=                         Slave status                         =\n"
printf "================================================================\n"
mysql --user="$SLAVE_USER" --password="$SLAVE_PASSWORD" --host="slave_db" -A \
  -e "SHOW SLAVE STATUS\G"