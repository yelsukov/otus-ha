#!/bin/bash

printf "\n\n================================================================\n"
printf "=                         Setup Shards                         =\n"
printf "================================================================\n"
for SHARD in "01" "02" "03"
do
  printf "[%s] Running replica for shard%s...\n" "$(date)" "${SHARD}"
  mongo --host "mongo${SHARD}Mstr" --port 27017 --eval "rs.initiate({_id: \"shard${SHARD}\", version: 1, members: [
  {_id:0, host:\"mongo${SHARD}Mstr:27017\"},
  {_id:1, host:\"mongo${SHARD}Repl:27017\"},
]})"
done

printf "\n\n================================================================\n"
printf "=                      Setup Config Server                     =\n"
printf "================================================================\n"
printf "[%s] Configuring the Config server...\n" "$(date)"
  mongo --host mongoConfig --port 27017 --eval "rs.initiate({_id:\"configServer\", configsvr:true, version:1, members:[
  { _id:0, host : \"mongoConfig:27017\" },
]})"

printf "\n\n================================================================\n"
printf "=                      Setup Router Server                     =\n"
printf "================================================================\n"

printf "[%s] Waiting for Router Server...\n" "$(date)"
RC=1
while [ $RC -eq 1 ]
do
    if ! mongo --host mongoRouter --port 27017 --eval 'quit(db.runCommand({ping:1}).ok?0:2)'  > /dev/null 2>&1;
    then
      sleep 1
    else
        RC=0
        echo "Router Server is alive!";
    fi
done

printf "\n[%s] Configuring the Router server...\n" "$(date)"
  mongo --host mongoRouter --port 27017 --eval "sh.addShard(\"shard01/mongo01Mstr:27017\");
sh.addShard(\"shard01/mongo01Repl:27017\");
sh.addShard(\"shard02/mongo02Mstr:27017\");
sh.addShard(\"shard02/mongo02Repl:27017\");
sh.addShard(\"shard03/mongo03Mstr:27017\");
sh.addShard(\"shard03/mongo03Repl:27017\");"

printf "\n\n[%s] Enabling sharding for \`%s\` database...\n" "$(date)" "${DB_NAME}"
  mongo --host mongoRouter --port 27017 --eval "sh.enableSharding(\"${DB_NAME}\")"

printf "\n[%s] Creating shared key for collection of messages...\n" "$(date)"
  mongo --host mongoRouter --port 27017 --eval "sh.shardCollection(\"${DB_NAME}.messages\", {\"cid\":\"hashed\", \"ts\":1})"

printf "\n[%s] Creating index for chat collection...\n" "$(date)"
  mongo --host mongoRouter --port 27017 --eval "db.getSiblingDB('${DB_NAME}'); db.chats.ensureIndex({'users': 1})"

printf "\n\n*******************************************************\n"
printf "********** Shared Cluster has been initiated **********\n"
printf "*******************************************************\n"
printf "[%s] Cluster status:\n" "$(date)"
  mongo --host mongoRouter --port 27017 --eval "sh.status()"