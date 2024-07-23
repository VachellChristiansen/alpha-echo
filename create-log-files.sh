#!/bin/bash

LOG_DIR="./logs"
LOG_FILES=("error.log" "warning.log" "task.log" "info.log" "misc.log")

if [ ! -d "$LOG_DIR" ]; then
    mkdir "$LOG_DIR"
    echo "Created directory: $LOG_DIR"
else
    echo "Directory already exists: $LOG_DIR"
fi

for LOG_FILE in "${LOG_FILES[@]}"; do
    FILE_PATH="$LOG_DIR/$LOG_FILE"
    if [ ! -f "$$FILE_PATH" ]; then
        touch "$FILE_PATH"
        echo "Created Log File: $FILE_PATH"
    else
        echo "Log File already exists: $FILE_PATH"
    fi
done

echo "Log directory and files setup completed."