#!/bin/bash

# Read the network configuration JSON file
configFile="./net_config.json"
dockerComposeFile="docker-compose.yml"

# Check if jq is installed
if ! command -v jq &> /dev/null
then
    echo "jq is not installed. Install it with 'sudo apt-get install jq'."
    exit 1
fi

# Loop through each node and create a Dockerfile
jq -c '.nodes[]' "$configFile" | while IFS= read -r node; do
    nodeName=$(jq -r '.name' <<< "$node")
    nodePort=$(jq -r '.port' <<< "$node")
    nodeAppPort=$(jq -r '.appPort' <<< "$node")
    nodeIdx=$(jq -r '.idx' <<< "$node")

    dockerfileName="Dockerfile.$nodeName"
    dockerfileContent=$(cat <<EOF
# Use an official Go image as a base
FROM golang:1.22

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code except the files specified in .dockerignore
COPY . .

# Build the application
RUN go build -o node_app ./src/main/node_app.go

# Expose the application ports
EXPOSE $nodePort
EXPOSE $nodeAppPort

# Command to run the application
CMD ["./node_app", "$nodeIdx", "$nodeAppPort", "net_config.json"]
EOF
)

    # Write the Dockerfile content to the file
    echo "$dockerfileContent" > "$dockerfileName"
done

# Generate Dockerfile for the app
dockerfileAppContent=$(cat <<EOF
# Use an official Go image as a base
FROM golang:1.22

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code except the files specified in .dockerignore
COPY . .

# Copy the final script
COPY merge-output.sh /merge-output.sh

# Set execute permission for the final script
RUN chmod +x /merge-output.sh

# Build the application
RUN go build -o application ./app.go

# Commands to run the application and, then, the final script
CMD ./application; bash /merge-output.sh
EOF
)

# Write the Dockerfile content to the file
echo "$dockerfileAppContent" > "Dockerfile.app"

# Start writing the docker-compose.yml
cat <<EOF > $dockerComposeFile
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.app
    container_name: chandy-lamport-app
    networks:
      - chandy-lamport-net
    volumes:
      - ./app/output:/app/output
    depends_on:
EOF

# Read nodes from the configuration file and add container names to the 'depends_on' section
nodes=$(jq -c '.nodes[]' $configFile)
for node in $nodes; do
  idx=$(echo $node | jq -r '.idx')
  echo "      - p$idx" >> $dockerComposeFile
done

# Add service definitions for processes P0, P1, etc.
for node in $nodes; do
  idx=$(echo $node | jq -r '.idx')
  name=$(echo $node | jq -r '.name')
  port=$(echo $node | jq -r '.port')
  appPort=$(echo $node | jq -r '.appPort')

  cat <<EOF >> $dockerComposeFile

  p$idx:
    build:
      context: .
      dockerfile: Dockerfile.$name
    container_name: chandy-lamport-p$idx
    networks:
      - chandy-lamport-net
    expose:
      - "$port"
      - "$appPort"
    volumes:
      - ./app/output:/app/output
EOF
done

# Add the network definition
cat <<EOF >> $dockerComposeFile

networks:
  chandy-lamport-net:
    driver: bridge
EOF

echo "Dockerfiles and docker-compose.yml have been generated successfully."
