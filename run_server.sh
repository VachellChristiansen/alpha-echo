#!/bin/bash

# Define the name of the executable or script to run
EXECUTABLE="your_go_executable"

# Function to find and kill the running server process
kill_server() {
  # Find the PID of the running server
  PID=$(lsof -t -i:YOUR_SERVER_PORT)

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
  # Adjust this line to use the correct command to run your server
  go run . &
  # Disown the process so it doesn't terminate when the script ends
  disown
}

# Main script logic
kill_server
start_server

echo "Server restarted successfully."
