#!/bin/bash

# -------------------------------
# Configuration
# -------------------------------
TIMEFRAME="6 months ago"
TOP_N=10

# -------------------------------
# Ensure we're in a Git repo
# -------------------------------
if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
    echo "This is not a Git repository. Please run inside a Git project."
    exit 1
fi

echo "Analyzing top $TOP_N contributors in: $(pwd)"
echo "Timeframe: Since $TIMEFRAME"

# -------------------------------
# Get total commits in timeframe
# -------------------------------
TOTAL_COMMITS=$(git log --since="$TIMEFRAME" --pretty=oneline | wc -l)

if [ "$TOTAL_COMMITS" -eq 0 ]; then
    echo "No commits found in the last 6 months."
    exit 0
fi

# -------------------------------
# Get and sort commits per author
# -------------------------------
echo
echo "Top $TOP_N Contributors by Commit Share:"
echo "------------------------------------------"

git shortlog -sne --since="$TIMEFRAME" \
| awk -v total="$TOTAL_COMMITS" '{
    count = $1
    sub($1 FS, "")
    printf "%s|%d|%.2f\n", $0, count, (count * 100) / total
}' \
| sort -t'|' -k2 -nr \
| head -n $TOP_N \
| awk -F'|' '{
    printf "%-30s %4d commits  (%5.2f%%)\n", $1, $2, $3
}'

echo "------------------------------------------"
echo "Total commits in timeframe: $TOTAL_COMMITS"
