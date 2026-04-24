import Database from "better-sqlite3";
import path from "path";
import os from "os";

// Cache database connections
const dbCache: Record<string, Database.Database> = {};

// Helper to get global DB
export function getGlobalDb(): Database.Database {
  const globalPath = path.join(os.homedir(), ".gemini", "antigravity", "global.db");

  if (!dbCache[globalPath]) {
    try {
      dbCache[globalPath] = new Database(globalPath, { readonly: true });
    } catch (err) {
      console.error(`Failed to connect to global DB at ${globalPath}:`, err);
      throw err;
    }
  }
  return dbCache[globalPath];
}

// Helper to get workspace context DB (returns null if unavailable)
export function getWorkspaceDb(): Database.Database | null {
  // Use WORKSPACE_PATH env var, or fallback to current directory if not provided
  const workspacePath = process.env.WORKSPACE_PATH || process.cwd();
  const contextPath = path.join(workspacePath, "context.db");

  if (!dbCache[contextPath]) {
    try {
      dbCache[contextPath] = new Database(contextPath, { readonly: true });
    } catch {
      // DB doesn't exist yet — return null so callers can return empty data
      return null;
    }
  }
  return dbCache[contextPath];
}

// Types for our DB models
export interface Task {
  task_id: string;
  description: string;
  status: string;
  notes: string;
  acceptance_criteria?: string;
  created_at: string;
  updated_at: string;
}

export interface Step {
  step_id: string;
  task_id: string;
  name: string;
  status: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface KnowledgeItem {
  ki_path: string;
  tactic_name: string;
  summary: string;
  decisions: string;
}

// Matches actual audit_logs schema
export interface AuditLog {
  id: number;
  timestamp: string;
  tool_name: string;
  request_payload: string;
  response_status: string;
  response_error: string | null;
}

export interface ActivityEvent {
  id: number;
  event_type: string;
  task_id: string;
  detail: string;
  created_at: string;
}

export interface DocReference {
  source_type: string;
  source_id: string;
  target_type: string;
  target_id: string;
  relation: string;
}
