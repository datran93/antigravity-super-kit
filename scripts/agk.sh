#!/bin/bash

# =============================================================================
# Antigravity Kit (agk) - Multi-Agent AI Management Tool
# =============================================================================

VERSION="2.0.0"

# --- Resolve Script Directory ---
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do
    DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
    SOURCE="$(readlink "$SOURCE")"
    [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
done
SCRIPT_DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

# --- Load Libraries ---
source "$SCRIPT_DIR/lib/agents.sh"
source "$SCRIPT_DIR/lib/transform.sh"

# --- Configuration ---
CACHE_DIR="$HOME/.antigravity/cache"
REPO_URL="git@github.com:datran93/antigravity-super-kit.git"
REPO_NAME="antigravity-kit"
SOURCE_AGENT_DIR="$CACHE_DIR/$REPO_NAME/.agents"
DEFAULT_AGENT="agy"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# =============================================================================
# Logging
# =============================================================================

log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1" >&2; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }

# =============================================================================
# Argument Parsing
# =============================================================================

# Parse --ai flags from arguments. Sets AGENTS array.
# Usage: parse_ai_args "$@"
parse_ai_args() {
    AGENTS=()
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --ai)
                shift
                if [[ -z "$1" || "$1" == --* ]]; then
                    log_error "Missing agent name after --ai"
                    exit 1
                fi
                if ! validate_agent "$1"; then
                    log_error "Unknown agent: $1"
                    echo "  Supported agents:"
                    list_agents
                    exit 1
                fi
                AGENTS+=("$1")
                shift
                ;;
            *)
                shift
                ;;
        esac
    done
    # Default to agy if no --ai specified
    [[ ${#AGENTS[@]} -eq 0 ]] && AGENTS=("$DEFAULT_AGENT")
}

# =============================================================================
# Core Functions
# =============================================================================

# Sync repository: compare HEAD, if different -> delete and re-clone
sync_repo() {
    mkdir -p "$CACHE_DIR" 2>/dev/null

    if [ -d "$CACHE_DIR/$REPO_NAME/.git" ]; then
        log_info "Checking for updates..."

        local local_head remote_head
        local_head=$(cd "$CACHE_DIR/$REPO_NAME" && git rev-parse HEAD 2>/dev/null)
        remote_head=$(git ls-remote "$REPO_URL" HEAD 2>/dev/null | cut -f1)

        if [ -z "$remote_head" ]; then
            log_warn "Cannot reach remote. Using cached version."
            return 0
        fi

        if [ "$local_head" = "$remote_head" ]; then
            log_success "Cache is up to date."
            return 0
        fi

        log_info "Updates available. Re-cloning..."
        rm -rf "$CACHE_DIR/$REPO_NAME" 2>/dev/null
    fi

    log_info "Cloning repository..."
    if ! git clone --depth 1 "$REPO_URL" "$CACHE_DIR/$REPO_NAME" 2>/dev/null; then
        log_error "Failed to clone repository."
        exit 1
    fi

    log_success "Repository synced."
}

# Get current source version (git short SHA)
get_source_version() {
    if [ -d "$CACHE_DIR/$REPO_NAME/.git" ]; then
        (cd "$CACHE_DIR/$REPO_NAME" && git rev-parse --short HEAD 2>/dev/null)
    else
        echo "unknown"
    fi
}

# Add a directory to git exclude
add_git_exclude() {
    local dir_name="$1"
    [ ! -d ".git" ] && return

    local exclude=".git/info/exclude"
    mkdir -p ".git/info" 2>/dev/null
    grep -q "^\\${dir_name}$" "$exclude" 2>/dev/null || echo "$dir_name" >> "$exclude"
}

# Remove a directory from git exclude
remove_git_exclude() {
    local dir_name="$1"
    [ ! -d ".git" ] && return

    local exclude=".git/info/exclude"
    if [ -f "$exclude" ] && grep -q "^\\${dir_name}$" "$exclude"; then
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "/^\\${dir_name}$/d" "$exclude"
        else
            sed -i "/^\\${dir_name}$/d" "$exclude"
        fi
    fi
}

# =============================================================================
# Commands
# =============================================================================

cmd_install() {
    parse_ai_args "$@"
    sync_repo

    local source_version
    source_version=$(get_source_version)

    local success_count=0
    for agent in "${AGENTS[@]}"; do
        if install_for_agent "$agent" "$SOURCE_AGENT_DIR" "."; then
            local target_dir_name
            target_dir_name=$(get_agent_field "$agent" "target_dir")
            local file_count
            file_count=$(count_files "./$target_dir_name")

            write_lock_file "." "$agent" "$source_version" "$file_count"
            log_success "Installed for $(get_agent_field "$agent" "name") → $target_dir_name/ ($file_count files)"
            ((success_count++))
        fi
    done

    if [[ $success_count -eq 0 ]]; then
        log_error "No agents were installed."
        exit 1
    fi
}

cmd_update() {
    parse_ai_args "$@"
    sync_repo

    local source_version
    source_version=$(get_source_version)

    for agent in "${AGENTS[@]}"; do
        if update_for_agent "$agent" "$SOURCE_AGENT_DIR" "."; then
            local target_dir_name
            target_dir_name=$(get_agent_field "$agent" "target_dir")
            local file_count
            file_count=$(count_files "./$target_dir_name")

            write_lock_file "." "$agent" "$source_version" "$file_count"
            log_success "Updated $(get_agent_field "$agent" "name") → $target_dir_name/ ($file_count files)"
        fi
    done
}

cmd_status() {
    sync_repo

    local lock_file=".agk.lock.json"
    if [ ! -f "$lock_file" ]; then
        log_error "No installations found. Run 'agk install --ai <agent>'."
        exit 1
    fi

    local source_version
    source_version=$(get_source_version)

    echo -e "${CYAN}Antigravity Kit Status${NC}"
    echo "  Current source: $source_version"
    echo ""
    read_lock_file "."
}

cmd_remove() {
    parse_ai_args "$@"

    for agent in "${AGENTS[@]}"; do
        local target_dir_name
        target_dir_name=$(get_agent_field "$agent" "target_dir")

        echo -e "${RED}WARNING: This will delete $target_dir_name/ for $agent.${NC}"
    done

    read -p "Continue? (y/N): " confirm
    [[ ! "$confirm" =~ ^[Yy]$ ]] && { log_info "Cancelled."; return; }

    for agent in "${AGENTS[@]}"; do
        if remove_for_agent "$agent" "."; then
            local target_dir_name
            target_dir_name=$(get_agent_field "$agent" "target_dir")
            remove_git_exclude "$target_dir_name"
            remove_from_lock_file "." "$agent"
            log_success "Removed $(get_agent_field "$agent" "name") ($target_dir_name/)"
        fi
    done
}

cmd_agents() {
    echo -e "${CYAN}Supported Agents${NC}"
    echo ""
    list_agents
    echo ""
}

cmd_info() {
    echo -e "${CYAN}Installed Agents${NC}"
    echo ""
    read_lock_file "."
    echo ""
}

show_help() {
    cat << EOF
${CYAN}Antigravity Kit v$VERSION${NC}

Usage: agk <command> [options]

Commands:
  install [--ai <agent>]   Install agent configuration (default: agy)
  update  [--ai <agent>]   Update agent to latest version
  status                   Check installation status
  remove  [--ai <agent>]   Remove agent configuration
  agents                   List all supported agents
  info                     Show current installations
  help                     Show this help

Options:
  --ai <agent>             Target agent (can be repeated for multi-install)
                           Default: agy (Antigravity native)

Examples:
  agk install                        # Install for Antigravity (default)
  agk install --ai claude            # Install for Claude Code
  agk install --ai claude --ai gemini # Install for multiple agents
  agk update --ai claude             # Update Claude installation
  agk remove --ai claude             # Remove Claude installation
  agk agents                         # List all supported agents

EOF
}

# =============================================================================
# Main
# =============================================================================

case "$1" in
    install) shift; cmd_install "$@" ;;
    update)  shift; cmd_update "$@" ;;
    status)  cmd_status ;;
    remove)  shift; cmd_remove "$@" ;;
    agents)  cmd_agents ;;
    info)    cmd_info ;;
    help|-h|--help|"") show_help ;;
    *) log_error "Unknown command: $1"; show_help; exit 1 ;;
esac
