#!/bin/bash

EXE_NAME="sonamusica-backend"
GO_FILES="."
GO_FLAGS="-v"

# Build the executable
go build $GO_FLAGS -o $EXE_NAME $GO_FILES

# Terminate on any error
if [ $? -ne 0 ]; then
  echo "Build failed, there were compile errors."
else
  # Run on no error
  echo "Build successful, running $EXE_NAME."
  ./$EXE_NAME
fi
