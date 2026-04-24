"use client";

import { useState, useEffect } from "react";
import { fetchTasks, fetchSteps, fetchKnowledgeItems, fetchAuditLogs } from "./actions";
import { Task, Step, KnowledgeItem, AuditLog } from "@/lib/db";
import { CheckCircle2, Circle, Clock, Network, ScrollText, Activity } from "lucide-react";

export default function Dashboard() {
  const [activeTab, setActiveTab] = useState<"tasks" | "knowledge" | "audit">("tasks");
  const [tasks, setTasks] = useState<Task[]>([]);
  const [selectedTaskSteps, setSelectedTaskSteps] = useState<Step[]>([]);
  const [knowledgeItems, setKnowledgeItems] = useState<KnowledgeItem[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [expandedKiId, setExpandedKiId] = useState<string | null>(null);

  useEffect(() => {
    fetchTasks().then(setTasks);
    fetchKnowledgeItems().then(setKnowledgeItems);
    fetchAuditLogs().then(setAuditLogs);
  }, []);

  const handleTaskClick = async (taskId: string) => {
    const steps = await fetchSteps(taskId);
    setSelectedTaskSteps(steps);
  };

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100 flex flex-col font-sans">
      <header className="border-b border-gray-800 bg-gray-900 p-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <h1 className="text-xl font-bold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
            Antigravity Kit Dashboard
          </h1>
          <nav className="flex space-x-1">
            <button
              onClick={() => setActiveTab("tasks")}
              className={`px-4 py-2 rounded-md flex items-center gap-2 text-sm font-medium transition-colors ${
                activeTab === "tasks" ? "bg-blue-600/20 text-blue-400" : "hover:bg-gray-800 text-gray-400"
              }`}
            >
              <CheckCircle2 size={16} /> Kanban Tasks
            </button>
            <button
              onClick={() => setActiveTab("knowledge")}
              className={`px-4 py-2 rounded-md flex items-center gap-2 text-sm font-medium transition-colors ${
                activeTab === "knowledge" ? "bg-purple-600/20 text-purple-400" : "hover:bg-gray-800 text-gray-400"
              }`}
            >
              <Network size={16} /> Knowledge Graph
            </button>
            <button
              onClick={() => setActiveTab("audit")}
              className={`px-4 py-2 rounded-md flex items-center gap-2 text-sm font-medium transition-colors ${
                activeTab === "audit" ? "bg-emerald-600/20 text-emerald-400" : "hover:bg-gray-800 text-gray-400"
              }`}
            >
              <Activity size={16} /> Audit Logs
            </button>
          </nav>
        </div>
      </header>

      <main className="flex-1 max-w-7xl mx-auto w-full p-6">
        {activeTab === "tasks" && (
          <div className="flex gap-6 h-full">
            {/* Task List */}
            <div className="w-1/3 flex flex-col gap-4">
              <h2 className="text-lg font-semibold flex items-center gap-2">
                <ScrollText size={18} className="text-gray-400" /> All Tasks
              </h2>
              <div className="flex flex-col gap-3 overflow-y-auto pr-2 pb-10">
                {tasks.map((task) => (
                  <div
                    key={task.task_id}
                    onClick={() => handleTaskClick(task.task_id)}
                    className="p-4 rounded-lg bg-gray-900 border border-gray-800 hover:border-gray-700 cursor-pointer transition-all"
                  >
                    <div className="flex justify-between items-start mb-2">
                      <h3 className="font-medium text-blue-400 truncate pr-2">{task.task_id}</h3>
                      <span
                        className={`text-xs px-2 py-1 rounded-full ${
                          task.status === "completed"
                            ? "bg-emerald-500/10 text-emerald-400"
                            : task.status === "in_progress"
                              ? "bg-blue-500/10 text-blue-400"
                              : "bg-gray-800 text-gray-400"
                        }`}
                      >
                        {task.status}
                      </span>
                    </div>
                    <p className="text-sm text-gray-400 line-clamp-2">{task.description}</p>
                  </div>
                ))}
                {tasks.length === 0 && <p className="text-gray-500 text-sm">No tasks found in project.</p>}
              </div>
            </div>

            {/* Task Details / Steps Kanban */}
            <div className="w-2/3 bg-gray-900/50 rounded-xl border border-gray-800 p-6 flex flex-col">
              {selectedTaskSteps.length > 0 ? (
                <>
                  <h2 className="text-lg font-semibold mb-6">Task Steps</h2>
                  <div className="grid grid-cols-2 gap-6 flex-1 min-h-0">
                    {/* Pending Column */}
                    <div className="flex flex-col gap-3 overflow-y-auto pr-2 pb-4">
                      <div className="flex items-center gap-2 mb-2 sticky top-0 bg-gray-900/50 backdrop-blur pt-1 pb-2 z-10">
                        <Circle size={16} className="text-blue-400" />
                        <h3 className="font-medium text-gray-300">Pending</h3>
                        <span className="text-xs bg-gray-800 text-gray-400 px-2 py-0.5 rounded-full">
                          {selectedTaskSteps.filter((s) => s.status !== "completed").length}
                        </span>
                      </div>
                      {selectedTaskSteps
                        .filter((s) => s.status !== "completed")
                        .map((step) => (
                          <div
                            key={step.step_id}
                            className="p-3 bg-gray-800/80 border border-gray-700 rounded-lg text-sm shrink-0"
                          >
                            <p className="text-gray-200">{step.name}</p>
                          </div>
                        ))}
                    </div>

                    {/* Completed Column */}
                    <div className="flex flex-col gap-3 overflow-y-auto pr-2 pb-4">
                      <div className="flex items-center gap-2 mb-2 sticky top-0 bg-gray-900/50 backdrop-blur pt-1 pb-2 z-10">
                        <CheckCircle2 size={16} className="text-emerald-400" />
                        <h3 className="font-medium text-gray-300">Completed</h3>
                        <span className="text-xs bg-gray-800 text-gray-400 px-2 py-0.5 rounded-full">
                          {selectedTaskSteps.filter((s) => s.status === "completed").length}
                        </span>
                      </div>
                      {selectedTaskSteps
                        .filter((s) => s.status === "completed")
                        .map((step) => (
                          <div
                            key={step.step_id}
                            className="p-3 bg-gray-800/40 border border-emerald-900/30 rounded-lg text-sm opacity-70 shrink-0"
                          >
                            <p className="text-emerald-100/70 line-through">{step.name}</p>
                          </div>
                        ))}
                    </div>
                  </div>
                </>
              ) : (
                <div className="flex-1 flex items-center justify-center text-gray-500">
                  Select a task to view its steps
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === "knowledge" && (
          <div className="flex flex-col gap-6">
            <h2 className="text-lg font-semibold flex items-center gap-2">
              <Network size={18} className="text-purple-400" /> Knowledge Items
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {knowledgeItems.map((ki) => (
                <div
                  key={ki.ki_path}
                  onClick={() => setExpandedKiId(expandedKiId === ki.ki_path ? null : ki.ki_path)}
                  className="p-5 rounded-xl bg-gray-900 border border-gray-800 hover:border-purple-500/30 transition-colors cursor-pointer"
                >
                  <h3 className="font-medium text-purple-300 mb-2">{ki.tactic_name}</h3>
                  <p className={`text-sm text-gray-400 mb-4 ${expandedKiId === ki.ki_path ? "" : "line-clamp-3"}`}>
                    {ki.summary}
                  </p>
                  {expandedKiId === ki.ki_path && ki.decisions && (
                    <div className="mb-4 p-3 bg-gray-950 rounded border border-gray-800">
                      <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">Decisions</h4>
                      <p className="text-sm text-gray-400">{ki.decisions}</p>
                    </div>
                  )}
                  <div className="text-xs text-gray-500 font-mono bg-gray-950 px-2 py-1 rounded inline-block">
                    {ki.ki_path.split("/").pop()}
                  </div>
                </div>
              ))}
              {knowledgeItems.length === 0 && <p className="text-gray-500">No knowledge items indexed.</p>}
            </div>
          </div>
        )}

        {activeTab === "audit" && (
          <div className="flex flex-col gap-6">
            <h2 className="text-lg font-semibold flex items-center gap-2">
              <Activity size={18} className="text-emerald-400" /> Global Audit Logs
            </h2>
            <div className="bg-gray-900 rounded-xl border border-gray-800 overflow-hidden">
              <table className="w-full text-sm text-left">
                <thead className="text-xs text-gray-400 bg-gray-950/50 uppercase border-b border-gray-800">
                  <tr>
                    <th className="px-4 py-3">Time</th>
                    <th className="px-4 py-3">Tool</th>
                    <th className="px-4 py-3">Status</th>
                    <th className="px-4 py-3">Duration</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-800">
                  {auditLogs.map((log) => (
                    <tr key={log.id} className="hover:bg-gray-800/50">
                      <td className="px-4 py-3 text-gray-500 whitespace-nowrap">
                        {new Date(log.timestamp).toLocaleTimeString()}
                      </td>
                      <td className="px-4 py-3 font-medium text-gray-300">{log.tool_name}</td>
                      <td className="px-4 py-3">
                        <span
                          className={`px-2 py-1 rounded-full text-xs ${
                            log.status === "success"
                              ? "bg-emerald-500/10 text-emerald-400"
                              : "bg-red-500/10 text-red-400"
                          }`}
                        >
                          {log.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-gray-500">{log.duration_ms}ms</td>
                    </tr>
                  ))}
                  {auditLogs.length === 0 && (
                    <tr>
                      <td colSpan={4} className="px-4 py-8 text-center text-gray-500">
                        No audit logs found.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
