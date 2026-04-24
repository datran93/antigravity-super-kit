"use server";

import { getGlobalDb, getWorkspaceDb, Task, Step, KnowledgeItem, AuditLog, ActivityEvent, DocReference } from "@/lib/db";
import type Database from "better-sqlite3";

// Helper: check if table exists before querying
function tableExists(db: Database.Database, tableName: string): boolean {
  const row = db.prepare("SELECT name FROM sqlite_master WHERE type='table' AND name=?").get(tableName) as { name: string } | undefined;
  return !!row;
}

export async function fetchTasks(): Promise<Task[]> {
  const db = getWorkspaceDb();
  if (!db) return [];
  return db.prepare("SELECT * FROM tasks ORDER BY updated_at DESC").all() as Task[];
}

export async function fetchSteps(taskId: string): Promise<Step[]> {
  const db = getWorkspaceDb();
  if (!db) return [];
  return db.prepare("SELECT * FROM steps WHERE task_id = ? ORDER BY created_at ASC").all(taskId) as Step[];
}

export async function fetchKnowledgeItems(): Promise<KnowledgeItem[]> {
  const db = getWorkspaceDb();
  if (!db) return [];
  if (!tableExists(db, "knowledge_fts")) return [];
  try {
    return db.prepare("SELECT tactic_name, ki_path, summary, decisions FROM knowledge_fts").all() as KnowledgeItem[];
  } catch {
    return [];
  }
}

export async function fetchAuditLogs(limit = 100): Promise<AuditLog[]> {
  const db = getGlobalDb();
  if (!tableExists(db, "audit_logs")) return [];
  try {
    return db.prepare("SELECT id, timestamp, tool_name, request_payload, response_status, response_error FROM audit_logs ORDER BY timestamp DESC LIMIT ?").all(limit) as AuditLog[];
  } catch {
    return [];
  }
}

export async function fetchActivityEvents(limit = 200): Promise<ActivityEvent[]> {
  const db = getWorkspaceDb();
  if (!db) return [];
  if (!tableExists(db, "activity_events")) return [];
  try {
    return db.prepare("SELECT * FROM activity_events ORDER BY created_at DESC LIMIT ?").all(limit) as ActivityEvent[];
  } catch {
    return [];
  }
}

export async function fetchDocReferences(): Promise<DocReference[]> {
  const db = getWorkspaceDb();
  if (!db) return [];
  if (!tableExists(db, "doc_references")) return [];
  try {
    return db.prepare("SELECT * FROM doc_references").all() as DocReference[];
  } catch {
    return [];
  }
}
