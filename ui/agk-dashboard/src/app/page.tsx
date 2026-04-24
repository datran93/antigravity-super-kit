"use client";

import { useState, useEffect } from "react";
import { fetchTasks, fetchSteps, fetchKnowledgeItems } from "./actions";
import { Task, Step, KnowledgeItem } from "@/lib/db";
import { CheckCircle2, Circle, ScrollText, Network, Clock } from "lucide-react";

export default function TasksPage() {
  const [activeTab, setActiveTab] = useState<"tasks" | "knowledge">("tasks");
  const [tasks, setTasks] = useState<Task[]>([]);
  const [selectedTaskSteps, setSelectedTaskSteps] = useState<Step[]>([]);
  const [knowledgeItems, setKnowledgeItems] = useState<KnowledgeItem[]>([]);
  const [expandedKiId, setExpandedKiId] = useState<string | null>(null);
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    const pollData = async () => {
      const newTasks = await fetchTasks();
      const newKnowledge = await fetchKnowledgeItems();

      if (isMounted) {
        setTasks(newTasks);
        setKnowledgeItems(newKnowledge);
      }

      if (selectedTaskId && isMounted) {
        const steps = await fetchSteps(selectedTaskId);
        if (isMounted) setSelectedTaskSteps(steps);
      }
    };

    pollData();
    const interval = setInterval(pollData, 3000);
    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, [selectedTaskId]);

  const handleTaskClick = async (taskId: string) => {
    setSelectedTaskId(taskId);
    const steps = await fetchSteps(taskId);
    setSelectedTaskSteps(steps);
  };

  const selectedTask = tasks.find((t) => t.task_id === selectedTaskId);
  const completedSteps = selectedTaskSteps.filter((s) => s.status === "completed");
  const pendingSteps = selectedTaskSteps.filter((s) => s.status !== "completed");

  return (
    <div className="h-full flex flex-col overflow-hidden">
      {/* Sub-tabs */}
      <div className="flex items-center gap-2 px-6 pt-5 pb-3 shrink-0 border-b border-gray-800/50">
        <button
          onClick={() => setActiveTab("tasks")}
          className={`px-4 py-2 rounded-lg flex items-center gap-2 text-sm font-medium transition-all ${
            activeTab === "tasks"
              ? "bg-blue-600/15 text-blue-400 shadow-sm shadow-blue-500/10"
              : "hover:bg-gray-800/60 text-gray-500"
          }`}
        >
          <ScrollText size={15} /> Tasks
        </button>
        <button
          onClick={() => setActiveTab("knowledge")}
          className={`px-4 py-2 rounded-lg flex items-center gap-2 text-sm font-medium transition-all ${
            activeTab === "knowledge"
              ? "bg-purple-600/15 text-purple-400 shadow-sm shadow-purple-500/10"
              : "hover:bg-gray-800/60 text-gray-500"
          }`}
        >
          <Network size={15} /> Knowledge
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === "tasks" && (
          <div className="flex h-full">
            {/* Task List — left panel */}
            <div className="w-80 lg:w-96 border-r border-gray-800/50 flex flex-col shrink-0">
              <div className="px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wider shrink-0">
                {tasks.length} task{tasks.length !== 1 ? "s" : ""}
              </div>
              <div className="flex-1 overflow-y-auto">
                {tasks.map((task) => (
                  <div
                    key={task.task_id}
                    onClick={() => handleTaskClick(task.task_id)}
                    className={`px-4 py-4 border-b border-gray-800/30 cursor-pointer transition-all ${
                      selectedTaskId === task.task_id
                        ? "bg-blue-950/30 border-l-2 border-l-blue-500"
                        : "hover:bg-gray-900/50 border-l-2 border-l-transparent"
                    }`}
                  >
                    <div className="flex items-start justify-between gap-2 mb-1.5">
                      <h3 className="font-medium text-sm text-gray-200 leading-tight">{task.task_id}</h3>
                      <span
                        className={`text-[10px] px-2 py-0.5 rounded-full font-medium shrink-0 ${
                          task.status === "completed"
                            ? "bg-emerald-500/15 text-emerald-400"
                            : task.status === "in_progress"
                              ? "bg-blue-500/15 text-blue-400"
                              : "bg-gray-800 text-gray-500"
                        }`}
                      >
                        {task.status === "in_progress" ? "active" : task.status}
                      </span>
                    </div>
                    <p className="text-xs text-gray-500 line-clamp-2 leading-relaxed">{task.description}</p>
                    <div className="flex items-center gap-2 mt-2 text-[10px] text-gray-600">
                      <Clock size={10} />
                      {new Date(task.updated_at).toLocaleDateString()}
                    </div>
                  </div>
                ))}
                {tasks.length === 0 && <div className="p-6 text-center text-gray-600 text-sm">No tasks found.</div>}
              </div>
            </div>

            {/* Task Detail — right panel */}
            <div className="flex-1 flex flex-col overflow-hidden">
              {selectedTaskId && selectedTask ? (
                <div className="flex-1 overflow-y-auto p-6">
                  {/* Header */}
                  <div className="mb-6">
                    <div className="flex items-center gap-3 mb-2">
                      <h2 className="text-lg font-bold text-gray-100">{selectedTask.task_id}</h2>
                      <span
                        className={`text-xs px-2.5 py-1 rounded-full font-medium ${
                          selectedTask.status === "completed"
                            ? "bg-emerald-500/15 text-emerald-400"
                            : "bg-blue-500/15 text-blue-400"
                        }`}
                      >
                        {selectedTask.status}
                      </span>
                    </div>
                    <p className="text-sm text-gray-400 leading-relaxed">{selectedTask.description}</p>
                  </div>

                  {/* Acceptance Criteria */}
                  {selectedTask.acceptance_criteria && (
                    <div className="bg-gray-900/80 rounded-lg p-4 border border-gray-800 mb-6">
                      <h3 className="text-xs font-semibold text-gray-500 uppercase mb-2 tracking-wider">
                        Acceptance Criteria
                      </h3>
                      <p className="text-sm text-gray-300 whitespace-pre-wrap leading-relaxed">
                        {selectedTask.acceptance_criteria}
                      </p>
                    </div>
                  )}

                  {/* Steps */}
                  {selectedTaskSteps.length > 0 ? (
                    <div>
                      <div className="flex items-center gap-3 mb-4">
                        <h3 className="text-sm font-semibold text-gray-300">Steps</h3>
                        <span className="text-xs text-gray-600">
                          {completedSteps.length}/{selectedTaskSteps.length} completed
                        </span>
                        {/* Progress bar */}
                        <div className="flex-1 h-1.5 bg-gray-800 rounded-full overflow-hidden max-w-xs">
                          <div
                            className="h-full bg-gradient-to-r from-emerald-500 to-emerald-400 rounded-full transition-all"
                            style={{ width: `${(completedSteps.length / selectedTaskSteps.length) * 100}%` }}
                          />
                        </div>
                      </div>

                      {/* Pending */}
                      {pendingSteps.length > 0 && (
                        <div className="mb-4">
                          <div className="flex items-center gap-2 mb-2">
                            <Circle size={12} className="text-blue-400" />
                            <span className="text-xs font-medium text-gray-500">Pending ({pendingSteps.length})</span>
                          </div>
                          <div className="space-y-1.5 ml-5">
                            {pendingSteps.map((step) => (
                              <div
                                key={step.step_id}
                                className="px-3 py-2.5 bg-gray-900/60 border border-gray-800 rounded-md text-sm text-gray-300"
                              >
                                {step.name}
                              </div>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* Completed */}
                      {completedSteps.length > 0 && (
                        <div>
                          <div className="flex items-center gap-2 mb-2">
                            <CheckCircle2 size={12} className="text-emerald-400" />
                            <span className="text-xs font-medium text-gray-500">
                              Completed ({completedSteps.length})
                            </span>
                          </div>
                          <div className="space-y-1.5 ml-5">
                            {completedSteps.map((step) => (
                              <div
                                key={step.step_id}
                                className="px-3 py-2.5 bg-gray-900/30 border border-gray-800/50 rounded-md text-sm text-gray-500 line-through"
                              >
                                {step.name}
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  ) : (
                    /* No steps — show notes */
                    <div>
                      <h3 className="text-xs font-semibold text-gray-500 uppercase mb-2 tracking-wider">Notes</h3>
                      <div className="bg-gray-900/60 rounded-lg p-4 border border-gray-800">
                        <p className="text-sm text-gray-400 whitespace-pre-wrap leading-relaxed">
                          {selectedTask.notes || "No notes recorded."}
                        </p>
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <div className="flex-1 flex items-center justify-center">
                  <div className="text-center">
                    <ScrollText size={40} className="mx-auto mb-3 text-gray-800" />
                    <p className="text-gray-600 text-sm">Select a task to view details</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === "knowledge" && (
          <div className="h-full overflow-y-auto p-6">
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
              {knowledgeItems.map((ki) => (
                <div
                  key={ki.ki_path}
                  onClick={() => setExpandedKiId(expandedKiId === ki.ki_path ? null : ki.ki_path)}
                  className="p-5 rounded-xl bg-gray-900 border border-gray-800 hover:border-purple-500/30 transition-colors cursor-pointer"
                >
                  <h3 className="font-medium text-purple-300 mb-2 text-sm">{ki.tactic_name}</h3>
                  <p
                    className={`text-sm text-gray-400 mb-3 whitespace-pre-wrap break-words leading-relaxed ${expandedKiId === ki.ki_path ? "" : "line-clamp-3"}`}
                  >
                    {ki.summary}
                  </p>
                  {expandedKiId === ki.ki_path && ki.decisions && (
                    <div className="mb-3 p-3 bg-gray-950 rounded border border-gray-800">
                      <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">Decisions</h4>
                      <p className="text-sm text-gray-400 whitespace-pre-wrap break-words">{ki.decisions}</p>
                    </div>
                  )}
                  <div className="text-xs text-gray-600 font-mono bg-gray-950 px-2 py-1 rounded inline-block">
                    {ki.ki_path.split("/").pop()}
                  </div>
                </div>
              ))}
              {knowledgeItems.length === 0 && (
                <p className="text-gray-600 col-span-full text-center py-12">No knowledge items indexed.</p>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
