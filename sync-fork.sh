#!/bin/bash

# Script to compare git tags and their merges
# Usage: ./compare-tags.sh <current_tag> <upstream_tag> <fork_branch>

set -euo pipefail
# set -x

# Check if we have the right number of arguments
if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <current_tag> <upstream_tag> <fork_branch>"
    echo "Example: $0 v0.46.16 v0.47.15 Agoric"
    exit 1
fi

CURRENT_TAG=$1
UPSTREAM_TAG=$2
FORK_BRANCH=$3
MERGE_FORK="merge-$FORK_BRANCH/$UPSTREAM_TAG"
DIFF_UPSTREAM="diff-$UPSTREAM_TAG/$CURRENT_TAG"
WORKSPACE="tag-comparison-workspace"

# Create a workspace directory
mkdir -p "$WORKSPACE"

# Function to check if a tag exists
check_tag() {
    if ! git rev-parse "$1" >/dev/null 2>&1; then
        echo "Error: Tag or branch '$1' does not exist"
        return 1
    fi
    return 0
}

# Function to cleanup
preserve_merge=true
cleanup() {
    echo "Cleaning up..."
    git checkout - >/dev/null 2>&1 || true
    $preserve_merge || git branch -D "$MERGE_FORK" >/dev/null 2>&1 || true
}

# Set up error handling
trap cleanup EXIT

# Verify tags exist
echo "Verifying tags and branch..."
check_tag "$CURRENT_TAG"
check_tag "$UPSTREAM_TAG"
check_tag "$FORK_BRANCH"

git branch --contains "$CURRENT_TAG" | grep -q "^. $FORK_BRANCH\$" || {
    echo "Error: fork branch '$FORK_BRANCH' does not contain current tag '$CURRENT_TAG'"
    exit 1
}

force() {
  cmd="$1"
  shift
  rm -f .git/index.lock
  if ! "$cmd" "$@"; then
    rm -f .git/index.lock
    "$cmd" "$@"
  fi
}

git_diff_add_rm() {
    a=$1 b=$2
    git diff --name-status "$a" "$b" | tee "$WORKSPACE/diff_add_rm.log" | \
    while read -r stat file renamed; do
      case $stat in
      D | R*) force git rm -f "$file" || true ;;
      A | M) force git add -f "$file" ;;
      esac
      case $stat in
      R*) force git add -f "$renamed" ;;
      esac
    done
}

git_diff_dont_delete() {
    a=$1
    git diff --name-status "$a" | tee "$WORKSPACE/diff-nodelete.log" | \
    while read -r stat file renamed; do
      case $stat in
      D) test ! -f "$file" || force git add -f "$file" ;;
      R*) force git add -f "$renamed" ;;
      esac
    done
}

# preserve_merge=false
if check_tag "$DIFF_UPSTREAM"; then
    echo "Using existing merge upstream branch $DIFF_UPSTREAM..."
else
    # Merge the upstream tag into the current tag.
    echo "Creating branch $DIFF_UPSTREAM..."
    git checkout -b "$DIFF_UPSTREAM" "$CURRENT_TAG"
    git merge -X theirs "$UPSTREAM_TAG" || true
    git checkout "$UPSTREAM_TAG" -- .
    git_diff_add_rm "$CURRENT_TAG" "$UPSTREAM_TAG"
    if test -f .git/MERGE_HEAD; then
      read -r -p "Please resolve any conflicts and press enter to continue..."
      git merge --continue
    else
      echo "No conflicts..."
    fi
fi

if check_tag "$MERGE_FORK"; then
    echo "Using existing merge fork branch $MERGE_FORK..."
    git checkout "$MERGE_FORK"
else
    # Merge the upstream tag into the fork branch.
    echo "Generating $MERGE_FORK..."
    git checkout -b "$MERGE_FORK" "$FORK_BRANCH"
    git merge "$DIFF_UPSTREAM" || true
    git_diff_dont_delete "$DIFF_UPSTREAM"
    if test -f .git/MERGE_HEAD; then
      read -r -p "Please resolve any conflicts and press enter to continue..."
      git merge --continue
    else
      echo "No conflicts..."
    fi
    preserve_merge=true
fi

echo "Starting comparison process..."
git config diff.renameLimit 999999

# Step 1: Check merge result
echo "Step 1: Checking merge result..."
git diff "$UPSTREAM_TAG" "$DIFF_UPSTREAM" > "$WORKSPACE/merge_upstream_diff.patch"

if [ -s "$WORKSPACE/merge_upstream_diff.patch" ]; then
    echo "Warning: Merge upstream result differs from upstream tag"
else
    echo "Merge result matches upstream tag"
fi

# Step 2: Create diff between current tag and Agoric branch
echo "Step 2: Creating diff between $CURRENT_TAG and $FORK_BRANCH..."
git diff "$CURRENT_TAG" "$FORK_BRANCH" > "$WORKSPACE/fork_changes.patch"

# Step 3: Create diff between upstream tag and final merge
echo "Step 3: Creating diff between $UPSTREAM_TAG and merged HEAD..."
git diff "$UPSTREAM_TAG" > "$WORKSPACE/final_changes.patch"

# Step 4: Compare the diffs
echo "Step 4: Comparing diffs..."
diff "$WORKSPACE/fork_changes.patch" "$WORKSPACE/final_changes.patch" > "$WORKSPACE/diff_comparison.txt" || true

# Generate summary
echo "
Summary:
--------
1. Merge upstream diff size: $(wc -l < "$WORKSPACE/merge_upstream_diff.patch") lines
2. Fork changes: $(wc -l < "$WORKSPACE/fork_changes.patch") lines
3. Final changes: $(wc -l < "$WORKSPACE/final_changes.patch") lines
4. Diff comparison: $(wc -l < "$WORKSPACE/diff_comparison.txt") lines

All files have been saved in the '$WORKSPACE' directory:
- merge_diff.patch: Differences between merge result and upstream
- fork_changes.patch: Original fork modifications
- final_changes.patch: Final state differences
- diff_comparison.txt: Comparison between the two sets of changes
"

if [ -s "$WORKSPACE/diff_comparison.txt" ]; then
    echo "Warning: Differences found between original fork changes and final state"
else
    echo "Success: No differences found between original fork changes and final state"
fi