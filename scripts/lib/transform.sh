#!/bin/bash

# =============================================================================
# transform.sh — Content Transformation Engine
# Converts canonical **/ content to agent-specific formats.
# Requires: agents.sh to be sourced first.
# =============================================================================

# -----------------------------------------------------------------------------
# transform_workflow_to_toml: Convert a Markdown workflow (with YAML frontmatter)
# to Gemini TOML format.
# Usage: transform_workflow_to_toml <input_file> <output_file> <args_placeholder>
# -----------------------------------------------------------------------------
transform_workflow_to_toml() {
    local input_file="$1" output_file="$2" args_placeholder="$3"
    python3 - "$input_file" "$output_file" "$args_placeholder" <<'PYEOF'
import sys, re

input_file = sys.argv[1]
output_file = sys.argv[2]
args_placeholder = sys.argv[3]

with open(input_file, 'r') as f:
    content = f.read()

# Parse YAML frontmatter
description = ""
body = content
if content.startswith("---"):
    parts = content.split("---", 2)
    if len(parts) >= 3:
        frontmatter = parts[1].strip()
        body = parts[2].strip()
        for line in frontmatter.split("\n"):
            if line.startswith("description:"):
                description = line.split(":", 1)[1].strip().strip('"').strip("'")

# Replace $ARGUMENTS with agent placeholder
body = body.replace("$ARGUMENTS", args_placeholder)

# Build TOML
toml_content = f'description = "{description}"\n\n{body}\n'

with open(output_file, 'w') as f:
    f.write(toml_content)
PYEOF
}

# -----------------------------------------------------------------------------
# transform_workflow_md: Convert a Markdown workflow for MD-based agents.
# Handles: extension change, args placeholder replacement, filename rename.
# Usage: transform_workflow_md <input_file> <output_file> <args_placeholder>
# -----------------------------------------------------------------------------
transform_workflow_md() {
    local input_file="$1" output_file="$2" args_placeholder="$3"

    # Copy and replace args placeholder
    if [[ "$args_placeholder" == '$ARGUMENTS' ]]; then
        cp "$input_file" "$output_file"
    else
        sed "s/\\\$ARGUMENTS/$args_placeholder/g" "$input_file" > "$output_file"
    fi
}

# -----------------------------------------------------------------------------
# transform_workflow_to_agent_md: Convert for GitHub Copilot (.agent.md format)
# Usage: transform_workflow_to_agent_md <input_file> <output_dir>
# -----------------------------------------------------------------------------
transform_workflow_to_agent_md() {
    local input_file="$1" output_dir="$2"
    local basename
    basename="$(basename "$input_file" .md)"
    local output_file="$output_dir/${basename}.agent.md"

    cp "$input_file" "$output_file"
}

# -----------------------------------------------------------------------------
# transform_workflow_to_skill_folder: Convert a workflow into a skill folder
# with SKILL.md for Codex-style agents.
# Usage: transform_workflow_to_skill_folder <input_file> <output_dir>
# -----------------------------------------------------------------------------
transform_workflow_to_skill_folder() {
    local input_file="$1" output_dir="$2"
    local basename
    basename="$(basename "$input_file" .md)"
    local skill_dir="$output_dir/$basename"

    mkdir -p "$skill_dir"
    cp "$input_file" "$skill_dir/SKILL.md"
}

# -----------------------------------------------------------------------------
# transform_workflows: Transform all workflows from source to target agent format.
# Usage: transform_workflows <agent_key> <source_dir> <target_dir>
# -----------------------------------------------------------------------------
transform_workflows() {
    local agent_key="$1" source_dir="$2" target_dir="$3"

    local wf_dir wf_format wf_ext args_placeholder
    wf_dir=$(get_agent_nested "$agent_key" "workflows" "dir")
    wf_format=$(get_agent_nested "$agent_key" "workflows" "format")
    wf_ext=$(get_agent_nested "$agent_key" "workflows" "extension")
    args_placeholder=$(get_agent_field "$agent_key" "args_placeholder")

    local output_dir="$target_dir/$wf_dir"
    mkdir -p "$output_dir"

    for src_file in "$source_dir"/workflows/*.md; do
        [ -f "$src_file" ] || continue
        local basename
        basename="$(basename "$src_file" .md)"

        case "$wf_format" in
            toml)
                transform_workflow_to_toml "$src_file" "$output_dir/${basename}${wf_ext}" "$args_placeholder"
                ;;
            skill_folder)
                transform_workflow_to_skill_folder "$src_file" "$output_dir"
                ;;
            md)
                if [[ "$wf_ext" == ".agent.md" ]]; then
                    transform_workflow_to_agent_md "$src_file" "$output_dir"
                else
                    transform_workflow_md "$src_file" "$output_dir/${basename}${wf_ext}" "$args_placeholder"
                fi
                ;;
        esac
    done
}

# -----------------------------------------------------------------------------
# transform_rules: Copy and adapt rules for the target agent.
# Usage: transform_rules <agent_key> <source_dir> <target_dir>
# -----------------------------------------------------------------------------
transform_rules() {
    local agent_key="$1" source_dir="$2" target_dir="$3"

    local rules_dir rules_ext
    rules_dir=$(get_agent_nested "$agent_key" "rules" "dir")
    rules_ext=$(get_agent_nested "$agent_key" "rules" "extension")

    local output_dir="$target_dir/$rules_dir"
    mkdir -p "$output_dir"

    # Get rename map as newline-separated "src:dst" pairs
    local rename_map_str
    rename_map_str=$(get_agent_rename_map "$agent_key")

    for src_file in "$source_dir"/rules/*.md; do
        [ -f "$src_file" ] || continue
        local basename
        basename="$(basename "$src_file")"

        # Look up rename: check if basename appears in rename_map
        local target_name="$basename"
        while IFS=':' read -r rn_src rn_dst; do
            if [[ -n "$rn_src" && "$rn_src" == "$basename" ]]; then
                target_name="$rn_dst"
                break
            fi
        done <<< "$rename_map_str"

        # Change extension if needed
        if [[ "$rules_ext" != ".md" ]]; then
            target_name="${target_name%.md}${rules_ext}"
        fi

        cp "$src_file" "$output_dir/$target_name"
    done
}

# -----------------------------------------------------------------------------
# copy_skills: Copy skill folders to target if agent supports skills.
# Usage: copy_skills <agent_key> <source_dir> <target_dir>
# -----------------------------------------------------------------------------
copy_skills() {
    local agent_key="$1" source_dir="$2" target_dir="$3"

    local skills_dir
    skills_dir=$(get_agent_nested "$agent_key" "skills" "dir")
    [ -z "$skills_dir" ] && return 0

    local output_dir="$target_dir/$skills_dir"
    if [ -d "$source_dir/skills" ]; then
        cp -R "$source_dir/skills" "$output_dir"
    fi
}

# -----------------------------------------------------------------------------
# copy_references: Copy reference files to target if agent supports them.
# Usage: copy_references <agent_key> <source_dir> <target_dir>
# -----------------------------------------------------------------------------
copy_references() {
    local agent_key="$1" source_dir="$2" target_dir="$3"

    local refs_dir
    refs_dir=$(get_agent_nested "$agent_key" "references" "dir")
    [ -z "$refs_dir" ] && return 0

    local output_dir="$target_dir/$refs_dir"
    if [ -d "$source_dir/references" ]; then
        cp -R "$source_dir/references" "$output_dir"
    fi
}

# -----------------------------------------------------------------------------
# install_for_agent: Full installation pipeline for a single agent.
# Usage: install_for_agent <agent_key> <source_agents_dir> <project_dir>
# -----------------------------------------------------------------------------
install_for_agent() {
    local agent_key="$1" source_dir="$2" project_dir="$3"

    local target_dir_name
    target_dir_name=$(get_agent_field "$agent_key" "target_dir")
    local full_target="$project_dir/$target_dir_name"

    # Check if target already exists
    if [ -d "$full_target" ]; then
        log_warn "$target_dir_name already exists. Use 'agk update --ai $agent_key' or remove it first."
        return 1
    fi

    # Verbatim agents (agy): direct copy
    if is_agent_verbatim "$agent_key"; then
        log_info "Installing $target_dir_name (verbatim copy)..."
        cp -R "$source_dir" "$full_target"
    else
        log_info "Installing $target_dir_name for $(get_agent_field "$agent_key" "name")..."
        mkdir -p "$full_target"
        transform_rules "$agent_key" "$source_dir" "$full_target"
        transform_workflows "$agent_key" "$source_dir" "$full_target"
        copy_skills "$agent_key" "$source_dir" "$full_target"
        copy_references "$agent_key" "$source_dir" "$full_target"
    fi

    return 0
}

# -----------------------------------------------------------------------------
# update_for_agent: Update an already installed agent.
# Usage: update_for_agent <agent_key> <source_agents_dir> <project_dir>
# -----------------------------------------------------------------------------
update_for_agent() {
    local agent_key="$1" source_dir="$2" project_dir="$3"

    local target_dir_name
    target_dir_name=$(get_agent_field "$agent_key" "target_dir")
    local full_target="$project_dir/$target_dir_name"

    if [ ! -d "$full_target" ]; then
        log_info "$target_dir_name not found. Installing..."
        install_for_agent "$agent_key" "$source_dir" "$project_dir"
        return $?
    fi

    log_info "Updating $target_dir_name..."
    rm -rf "$full_target"

    if is_agent_verbatim "$agent_key"; then
        cp -R "$source_dir" "$full_target"
    else
        mkdir -p "$full_target"
        transform_rules "$agent_key" "$source_dir" "$full_target"
        transform_workflows "$agent_key" "$source_dir" "$full_target"
        copy_skills "$agent_key" "$source_dir" "$full_target"
        copy_references "$agent_key" "$source_dir" "$full_target"
    fi

    return 0
}

# -----------------------------------------------------------------------------
# remove_for_agent: Remove an installed agent's files.
# Usage: remove_for_agent <agent_key> <project_dir>
# -----------------------------------------------------------------------------
remove_for_agent() {
    local agent_key="$1" project_dir="$2"

    local target_dir_name
    target_dir_name=$(get_agent_field "$agent_key" "target_dir")
    local full_target="$project_dir/$target_dir_name"

    if [ ! -d "$full_target" ]; then
        log_warn "$target_dir_name not found for $agent_key."
        return 1
    fi

    rm -rf "$full_target"
    return 0
}

# -----------------------------------------------------------------------------
# write_lock_file: Write or update .agk.lock.json
# Usage: write_lock_file <project_dir> <agent_key> <source_version> <file_count>
# -----------------------------------------------------------------------------
write_lock_file() {
    local project_dir="$1" agent_key="$2" source_version="$3" file_count="$4"
    local lock_file="$project_dir/.agk.lock.json"
    local target_dir_name
    target_dir_name=$(get_agent_field "$agent_key" "target_dir")
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    python3 - "$lock_file" "$agent_key" "$target_dir_name" "$source_version" "$file_count" "$timestamp" <<'PYEOF'
import json, sys, os

lock_file = sys.argv[1]
agent_key = sys.argv[2]
target_dir = sys.argv[3]
source_version = sys.argv[4]
file_count = int(sys.argv[5])
timestamp = sys.argv[6]

# Load existing or create new
if os.path.exists(lock_file):
    with open(lock_file) as f:
        lock = json.load(f)
else:
    lock = {"schema_version": "1.0", "agents": {}}

lock["source_version"] = source_version
lock["agents"][agent_key] = {
    "target_dir": target_dir,
    "files_installed": file_count,
    "installed_at": timestamp
}

with open(lock_file, 'w') as f:
    json.dump(lock, f, indent=2)
    f.write('\n')
PYEOF
}

# -----------------------------------------------------------------------------
# remove_from_lock_file: Remove an agent from .agk.lock.json
# Usage: remove_from_lock_file <project_dir> <agent_key>
# -----------------------------------------------------------------------------
remove_from_lock_file() {
    local project_dir="$1" agent_key="$2"
    local lock_file="$project_dir/.agk.lock.json"

    [ ! -f "$lock_file" ] && return 0

    python3 - "$lock_file" "$agent_key" <<'PYEOF'
import json, sys, os

lock_file = sys.argv[1]
agent_key = sys.argv[2]

with open(lock_file) as f:
    lock = json.load(f)

lock["agents"].pop(agent_key, None)

if not lock["agents"]:
    os.remove(lock_file)
else:
    with open(lock_file, 'w') as f:
        json.dump(lock, f, indent=2)
        f.write('\n')
PYEOF
}

# -----------------------------------------------------------------------------
# read_lock_file: Print lock file contents (for cmd_info)
# Usage: read_lock_file <project_dir>
# -----------------------------------------------------------------------------
read_lock_file() {
    local project_dir="$1"
    local lock_file="$project_dir/.agk.lock.json"

    if [ ! -f "$lock_file" ]; then
        echo "  No installations found. Run 'agk install --ai <agent>'."
        return 1
    fi

    python3 - "$lock_file" <<'PYEOF'
import json, sys

with open(sys.argv[1]) as f:
    lock = json.load(f)

print(f"  Source version: {lock.get('source_version', 'unknown')}")
print()
for key, info in lock.get("agents", {}).items():
    tdir = info.get("target_dir", "?")
    count = info.get("files_installed", "?")
    ts = info.get("installed_at", "?")
    print(f"  {key:12s}  {tdir:15s}  {count} files  ({ts})")
PYEOF
}

# -----------------------------------------------------------------------------
# count_files: Count files in a directory (for lock file metadata)
# Usage: count_files <dir>
# -----------------------------------------------------------------------------
count_files() {
    find "$1" -type f 2>/dev/null | wc -l | tr -d ' '
}
