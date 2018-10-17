#!/usr/bin/env bash
# Listen to a batch of directories for updates
# look for marker to detect a complete pod run
# push the directory contents (github?)

# Todo move to go
apt-get update
apt-get install -y inotify-tools
while true
do
  # TODO:
  # Monitor pods and look for pod specific paths and a signature file with contents with validation
  # Instead of inotify watch kubernetes pods status on successful completion
  # Extract the path information for a pod and inspect for completion.
  # Commit the output to datamon backend
  inotifywait -mr /data/output/ -e create -e moved_to| while read path action file; do if [[ $file = *done* ]]; then echo "output done"; fi; done
  cat /data/output/repo/data
done
