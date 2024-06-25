#!/bin/bash

EXECUTABLE="alpha-echo"

# Function to find and kill the running server process
kill_server() {
  PID=$(lsof -t -i:49153)

  if [ -n "$PID" ]; then
    echo "Killing existing server process with PID: $PID"
    kill -9 $PID
  else
    echo "No running server process found."
  fi
}

# Function to start the server
start_server() {
  echo "Starting the server..."
  go run . &
  disown
}

kill_server
start_server

echo "Server restarted successfully."
