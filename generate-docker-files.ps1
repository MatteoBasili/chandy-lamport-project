# Read the network configuration JSON file
$configFile = Get-Content -Raw -Path ".\net_config.json" | ConvertFrom-Json

# Loop through each node and create a Dockerfile
foreach ($node in $configFile.nodes) {
    $dockerfileName = "Dockerfile." + $node.name
    $dockerfileContent = @"
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
EXPOSE $($node.port)
EXPOSE $($node.appPort)

# Command to run the application
CMD ["./node_app", "$($node.idx)", "$($node.appPort)", "net_config.json"]
"@
    # Write the Dockerfile content to the file
    Set-Content -Path $dockerfileName -Value $dockerfileContent
}

# Generate Dockerfile for the app
$dockerfileAppContent = @"
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
RUN go build -o app ./app.go

# Command to run the application
CMD ["./app"]
"@
# Write the Dockerfile content to the file
Set-Content -Path "Dockerfile.app" -Value $dockerfileAppContent

# Create the initial content of the docker-compose.yml file
$dockerComposeContent = 
"services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.app
    container_name: chandy-lamport-app
    networks:
      - chandy-lamport-net
    depends_on:"

# Add dependencies for the app
foreach ($node in $configFile.nodes) {
    $dockerComposeContent += "
      - p$($node.idx)"
}

$dockerComposeContent +="
"

# Loop through each node and append the service definition to the docker-compose.yml content
foreach ($node in $configFile.nodes) {
    $dockerComposeContent += "
  p$($node.idx):
    build:
      context: .
      dockerfile: Dockerfile.$($node.name)
    container_name: chandy-lamport-p$($node.idx)
    networks:
      - chandy-lamport-net
    expose:
      - ""$($node.port)""
      - ""$($node.appPort)""
"
}

# Add the networks section
$dockerComposeContent += "
networks:
  chandy-lamport-net:
    driver: bridge"

# Replace tab characters with spaces in the YAML content
$dockerComposeContent = $dockerComposeContent -replace "`t", "    "

# Write the docker-compose.yml content to a file
Set-Content -Path "docker-compose.yml" -Value $dockerComposeContent

Write-Output "Dockerfiles and docker-compose.yml have been generated successfully."
