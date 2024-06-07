#!/bin/bash

# Check if the configuration file exists
CONFIG_FILE="net_config.json"
if [[ ! -f $CONFIG_FILE ]]; then
  echo "Configuration file $CONFIG_FILE not found!"
  exit 1
fi

# Read the content of the JSON file using jq
NODES=$(jq -c '.nodes[]' $CONFIG_FILE)

# Iterate over each node and launch the corresponding application
for NODE in $NODES; do
  IDX=$(echo $NODE | jq -r '.idx')
  APP_PORT=$(echo $NODE | jq -r '.appPort')
  
  echo "Launching application for node $IDX on port $APP_PORT..."
  go run ./src/main/node_app.go $IDX $APP_PORT $CONFIG_FILE &
done

