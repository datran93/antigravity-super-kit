"use server";

import { getGlobalDb, getWorkspaceDb, Task, Step, KnowledgeItem, AuditLog } from "@/lib/db";

export async function fetchTasks(): Promise<Task[]> {
  const db = getWorkspaceDb();
  return db.prepare("SELECT * FROM tasks ORDER BY updated_at DESC").all() as Task[];
}

export async function fetchSteps(taskId: string): Promise<Step[]> {
  const db = getWorkspaceDb();
  return db.prepare("SELECT * FROM steps WHERE task_id = ? ORDER BY created_at ASC").all(taskId) as Step[];
}

export async function fetchKnowledgeItems(): Promise<KnowledgeItem[]> {
  const db = getWorkspaceDb();
  // We need to check if the table exists first as the workspace might not have any knowledge items indexed yet
  try {
    return db.prepare("SELECT tactic_name, ki_path, summary, decisions FROM knowledge_fts").all() as KnowledgeItem[];
  } catch (err) {
    console.error("Failed to fetch knowledge items, table might not exist yet:", err);
    return [];
  }
}

export async function fetchAuditLogs(limit = 100): Promise<AuditLog[]> {
  const db = getGlobalDb();
  try {
    return db.prepare("SELECT * FROM audit_logs ORDER BY timestamp DESC LIMIT ?").all(limit) as AuditLog[];
  } catch (err) {
    console.error("Failed to fetch audit logs, table might not exist yet:", err);
    return [];
  }
}
