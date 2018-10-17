#!/usr/bin/env bash
# Read from input path and output to folder based on environment variables
cat /data/input/repo/file >> /data/output/repo/file
echo "Processing done" | tee -a /data/output/repo/file
