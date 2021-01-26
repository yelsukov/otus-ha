#!/bin/bash

export ORCHESTRATOR_API="http://localhost:3000/api"

printf "Orchestrator discovering master db...\n"
/usr/local/orchestrator/resources/bin/orchestrator-client -c discover -i master_db

printf "\nThe following topology is available:\n"
CLUSTER="$(/usr/local/orchestrator/resources/bin/orchestrator-client -c clusters)"
/usr/local/orchestrator/resources/bin/orchestrator-client -c topology -a "$CLUSTER"
printf "\nOrchestrator discovery finished! Cluster name is: %s\n" "$CLUSTER"