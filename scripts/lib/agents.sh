#!/bin/bash

# =============================================================================
# agents.sh — Agent Configuration Loader
# Parses agents.json to provide agent config access from shell.
# =============================================================================

# Locate agents.json relative to this library
_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
_AGENTS_JSON="$_LIB_DIR/../../agents.json"

# -----------------------------------------------------------------------------
# _agent_py: Run a Python script against agents.json, passing args via sys.argv
# Usage: _agent_py <script_name> [args...]
# The script receives: sys.argv[1]=agents.json path, sys.argv[2..]=user args
# data is pre-loaded as a global.
# -----------------------------------------------------------------------------
_agent_py() {
    local script="$1"
    shift
    python3 - "$_AGENTS_JSON" "$@" <<PYEOF
import json, sys

with open(sys.argv[1]) as f:
    data = json.load(f)
agents = data["agents"]

${script}
PYEOF
}

# -----------------------------------------------------------------------------
# list_agents: Print all supported agent keys with name and target_dir
# -----------------------------------------------------------------------------
list_agents() {
    _agent_py '
for key in sorted(agents):
    a = agents[key]
    name = a["name"]
    tdir = a["target_dir"]
    print(f"  {key:12s}  {name:25s}  -> {tdir}")
'
}

# -----------------------------------------------------------------------------
# validate_agent: Check if an agent key exists. Returns 0/1.
# Usage: validate_agent "claude"
# -----------------------------------------------------------------------------
validate_agent() {
    local agent_key="$1"
    _agent_py 'sys.exit(0 if sys.argv[2] in agents else 1)' "$agent_key"
}

# -----------------------------------------------------------------------------
# get_agent_field: Get a top-level field from an agent config.
# Usage: get_agent_field "claude" "target_dir"
# -----------------------------------------------------------------------------
get_agent_field() {
    local agent_key="$1" field="$2"
    _agent_py '
key, field = sys.argv[2], sys.argv[3]
a = agents.get(key, {})
v = a.get(field, "")
if isinstance(v, bool):
    print("true" if v else "false")
elif isinstance(v, dict) or isinstance(v, list):
    print(json.dumps(v))
elif v is None:
    print("")
else:
    print(v)
' "$agent_key" "$field"
}

# -----------------------------------------------------------------------------
# get_agent_nested: Get a nested field from an agent config.
# Usage: get_agent_nested "claude" "rules" "dir"
# -----------------------------------------------------------------------------
get_agent_nested() {
    local agent_key="$1" section="$2" field="$3"
    _agent_py '
key, section, field = sys.argv[2], sys.argv[3], sys.argv[4]
a = agents.get(key, {})
s = a.get(section)
if s is None or not isinstance(s, dict):
    print("")
else:
    v = s.get(field, "")
    if isinstance(v, dict):
        print(json.dumps(v))
    elif v is None:
        print("")
    else:
        print(v)
' "$agent_key" "$section" "$field"
}

# -----------------------------------------------------------------------------
# get_agent_rename_map: Get the rules rename map as "src:dst" lines
# Usage: get_agent_rename_map "claude"  →  "GEMINI.md:CLAUDE.md"
# -----------------------------------------------------------------------------
get_agent_rename_map() {
    local agent_key="$1"
    _agent_py '
key = sys.argv[2]
a = agents.get(key, {})
rules = a.get("rules")
if rules and isinstance(rules, dict):
    rename = rules.get("rename", {})
    for src, dst in rename.items():
        print(f"{src}:{dst}")
' "$agent_key"
}

# -----------------------------------------------------------------------------
# is_agent_verbatim: Check if agent uses verbatim copy (agy).
# Usage: is_agent_verbatim "agy"  →  returns 0 if true
# -----------------------------------------------------------------------------
is_agent_verbatim() {
    local agent_key="$1"
    local val
    val=$(get_agent_field "$agent_key" "verbatim")
    [[ "$val" == "true" ]]
}
