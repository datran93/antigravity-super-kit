#!/bin/bash

# sync-skills.sh
# Synchronizes local .agent/skills with the antigravity-awesome-skills repository.
# Only updates skills that are already present in the local directory and have changes.

# --- Configuration ---
LOCAL_SKILLS_DIR="$(pwd)/.agent/skills"
SOURCE_SKILLS_DIR="/Users/datran/LearnDev/antigravity-awesome-skills/skills"
DRY_RUN=false
FORCE=false

# --- Argument Parsing ---
for arg in "$@"; do
  if [ "$arg" == "--dry-run" ]; then
    DRY_RUN=true
    echo "🔍 DRY RUN: No files will be changed."
  elif [ "$arg" == "--force" ]; then
    FORCE=true
  fi
done

# --- Validation ---
if [ ! -d "$LOCAL_SKILLS_DIR" ]; then
  echo "❌ Error: Local skills directory not found at $LOCAL_SKILLS_DIR"
  exit 1
fi

if [ ! -d "$SOURCE_SKILLS_DIR" ]; then
  echo "❌ Error: Source skills directory not found at $SOURCE_SKILLS_DIR"
  exit 1
fi

echo "🔄 Checking for updates in $LOCAL_SKILLS_DIR..."

# --- Sync Logic ---
sync_count=0
updated_count=0
no_change_count=0

for skill_path in "$LOCAL_SKILLS_DIR"/*; do
  if [ -d "$skill_path" ]; then
    skill_name=$(basename "$skill_path")
    source_skill="$SOURCE_SKILLS_DIR/$skill_name"

    if [ -d "$source_skill" ]; then
      sync_count=$((sync_count + 1))

      # Quick check for changes using rsync dry-run itemized output
      # We exclude common noise like .DS_Store
      CHANGES=$(rsync -ni -av --delete --exclude=".DS_Store" "$source_skill/" "$skill_path/" | grep -v "^\." | wc -l | xargs)

      if [ "$CHANGES" -gt 0 ] || [ "$FORCE" = true ]; then
        if [ "$DRY_RUN" = true ]; then
          echo "  [PENDING] $skill_name ($CHANGES changes detected)"
        else
          echo "  [UPDATING] $skill_name ($CHANGES changes detected)..."
          rsync -av --delete --exclude=".DS_Store" "$source_skill/" "$skill_path/" > /dev/null
          updated_count=$((updated_count + 1))
        fi
      else
        no_change_count=$((no_change_count + 1))
      fi
    # else: Skill purely local, ignore as per requirement
    fi
  fi
done

echo ""
if [ "$DRY_RUN" = true ]; then
  echo "✅ Dry run complete. Found $sync_count local skills ($((sync_count - no_change_count)) need update)."
else
  echo "✅ Finished! Updated $updated_count skills. $no_change_count skills were already up to date."
fi
