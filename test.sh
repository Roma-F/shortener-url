#!/bin/bash

# Define the log file
LOG_FILE="test_log_file.txt"

go build -ldflags '-s -w -X main.buildVersion=1.0.0 -X main.buildDate=2023-01-23 -X main.buildCommit=0c2fs'  -o cmd/shortener/shortener cmd/shortener/main.go

# Run your test commands, capturing output to the log file
./shortenertestbeta -test.v -test.run=^TestIteration1$ -binary-path=cmd/shortener/shortener > "$LOG_FILE" 2>&1
./shortenertestbeta -test.v -test.run=^TestIteration2$ -source-path=.  > "$LOG_FILE" 2>&1
./shortenertestbeta -test.v -test.run=^TestIteration3$ -source-path=. > "$LOG_FILE" 2>&1
./shortenertestbeta -test.v -test.run=^TestIteration4$ -binary-path=cmd/shortener/shortener -server-port=8083 > "$LOG_FILE" 2>&1
./shortenertestbeta -test.v -test.run=^TestIteration5$ -binary-path=cmd/shortener/shortener -server-port=8083 > "$LOG_FILE" 2>&1
./shortenertestbeta -test.v -test.run=^TestIteration6$ -binary-path=cmd/shortener/shortener  -source-path=. -server-port=8083 > "$LOG_FILE" 2>&1

# Process the log file to find failed tests
while read line; do
  # Check if the line indicates a failed test
  if [[ "$line" =~ "FAIL" ]]; then
    echo "$line"
  fi
done < "$LOG_FILE"