"use client";

import { useState, useEffect, useMemo } from "react";
import { fetchActivityEvents, fetchAuditLogs } from "../actions";
import { ActivityEvent, AuditLog } from "@/lib/db";
import { Activity, Filter, Clock, Zap, Shield } from "lucide-react";

const EVENT_COLORS: Record<string, { bg: string; text: string; icon: string }> = {
  step_completed: { bg: "bg-emerald-500/10", text: "text-emerald-400", icon: "✅" },
  task_initialized: { bg: "bg-blue-500/10", text: "text-blue-400", icon: "🚀" },
  ki_created: { bg: "bg-purple-500/10", text: "text-purple-400", icon: "🧠" },
  intent_declared: { bg: "bg-amber-500/10", text: "text-amber-400", icon: "🔒" },
  failure_recorded: { bg: "bg-red-500/10", text: "text-red-400", icon: "❌" },
};

const DEFAULT_COLOR = { bg: "bg-gray-500/10", text: "text-gray-400", icon: "📋" };

export default function ActivityPage() {
  const [events, setEvents] = useState<ActivityEvent[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [activeView, setActiveView] = useState<"events" | "audit">("audit");
  const [typeFilter, setTypeFilter] = useState<string>("all");

  useEffect(() => {
    let isMounted = true;

    const poll = async () => {
      const [ev, logs] = await Promise.all([fetchActivityEvents(), fetchAuditLogs()]);
      if (isMounted) {
        setEvents(ev);
        setAuditLogs(logs);
        // Auto-switch to events if they exist
        if (ev.length > 0 && logs.length === 0) setActiveView("events");
      }
    };

    poll();
    const interval = setInterval(poll, 5000);
    return () => { isMounted = false; clearInterval(interval); };
  }, []);

  const eventTypes = useMemo(() => {
    const types = new Set(events.map((e) => e.event_type));
    return ["all", ...Array.from(types).sort()];
  }, [events]);

  const filtered = useMemo(() => {
    if (typeFilter === "all") return events;
    return events.filter((e) => e.event_type === typeFilter);
  }, [events, typeFilter]);

  return (
    <div className="h-full flex flex-col overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-6 pt-5 pb-3 border-b border-gray-800/50 shrink-0">
        <h1 className="text-lg font-bold flex items-center gap-2.5">
          <Activity size={20} className="text-emerald-400" />
          Activity Feed
        </h1>
        <div className="flex items-center gap-1.5 bg-gray-900 rounded-lg p-1 border border-gray-800">
          <button
            onClick={() => setActiveView("events")}
            className={`px-3 py-1.5 rounded-md text-xs font-medium flex items-center gap-1.5 transition-all ${
              activeView === "events"
                ? "bg-emerald-600/15 text-emerald-400 shadow-sm"
                : "text-gray-500 hover:text-gray-300"
            }`}
          >
            <Zap size={12} /> Events {events.length > 0 && <span className="bg-gray-800 px-1.5 py-0.5 rounded-full text-[10px]">{events.length}</span>}
          </button>
          <button
            onClick={() => setActiveView("audit")}
            className={`px-3 py-1.5 rounded-md text-xs font-medium flex items-center gap-1.5 transition-all ${
              activeView === "audit"
                ? "bg-emerald-600/15 text-emerald-400 shadow-sm"
                : "text-gray-500 hover:text-gray-300"
            }`}
          >
            <Shield size={12} /> Audit {auditLogs.length > 0 && <span className="bg-gray-800 px-1.5 py-0.5 rounded-full text-[10px]">{auditLogs.length}</span>}
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-hidden">
        {activeView === "events" && (
          <div className="h-full flex flex-col">
            {/* Filter bar */}
            {events.length > 0 && (
              <div className="flex items-center gap-3 px-6 py-3 border-b border-gray-800/30 shrink-0">
                <Filter size={12} className="text-gray-600" />
                <div className="flex gap-1.5 flex-wrap">
                  {eventTypes.map((type) => (
                    <button
                      key={type}
                      onClick={() => setTypeFilter(type)}
                      className={`px-2.5 py-1 rounded-full text-[10px] font-medium transition-colors ${
                        typeFilter === type
                          ? "bg-emerald-600/20 text-emerald-400 border border-emerald-500/30"
                          : "bg-gray-900 text-gray-500 border border-gray-800 hover:border-gray-700"
                      }`}
                    >
                      {type === "all" ? "All" : type.replace(/_/g, " ")}
                    </button>
                  ))}
                </div>
                <span className="text-[10px] text-gray-600 ml-auto">{filtered.length} events</span>
              </div>
            )}

            {/* Timeline */}
            <div className="flex-1 overflow-y-auto px-6 py-4">
              {filtered.length > 0 ? (
                <div className="relative">
                  <div className="absolute left-3 top-0 bottom-0 w-px bg-gray-800" />
                  {filtered.map((event) => {
                    const color = EVENT_COLORS[event.event_type] || DEFAULT_COLOR;
                    return (
                      <div key={event.id} className="relative pl-9 pb-4 group">
                        <div className={`absolute left-1 top-1.5 w-5 h-5 rounded-full flex items-center justify-center text-[10px] ${color.bg} border border-gray-800 group-hover:scale-110 transition-transform`}>
                          {color.icon}
                        </div>
                        <div className="bg-gray-900 border border-gray-800 rounded-lg p-3.5 hover:border-gray-700 transition-colors">
                          <div className="flex items-center justify-between mb-1.5">
                            <div className="flex items-center gap-2">
                              <span className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${color.bg} ${color.text}`}>
                                {event.event_type.replace(/_/g, " ")}
                              </span>
                              {event.task_id && <span className="text-[10px] text-gray-600 font-mono">{event.task_id}</span>}
                            </div>
                            <span className="text-[10px] text-gray-600 flex items-center gap-1">
                              <Clock size={9} />
                              {new Date(event.created_at).toLocaleString()}
                            </span>
                          </div>
                          <p className="text-xs text-gray-400 leading-relaxed">{event.detail}</p>
                        </div>
                      </div>
                    );
                  })}
                </div>
              ) : (
                <div className="flex items-center justify-center h-full">
                  <div className="text-center">
                    <Zap size={36} className="mx-auto mb-3 text-gray-800" />
                    <p className="text-gray-600 text-sm">No activity events recorded yet.</p>
                    <p className="text-gray-700 text-xs mt-1">Events appear as agents run MCP tools.</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {activeView === "audit" && (
          <div className="h-full overflow-y-auto">
            <table className="w-full text-sm text-left">
              <thead className="text-[10px] text-gray-500 uppercase tracking-wider border-b border-gray-800 sticky top-0 bg-gray-950 z-10">
                <tr>
                  <th className="px-6 py-3 font-medium">Time</th>
                  <th className="px-6 py-3 font-medium">Tool</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                  <th className="px-6 py-3 font-medium">Details</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-800/50">
                {auditLogs.map((log) => (
                  <tr key={log.id} className="hover:bg-gray-900/50 transition-colors">
                    <td className="px-6 py-3 text-gray-500 whitespace-nowrap text-xs">
                      {new Date(log.timestamp).toLocaleTimeString()}
                    </td>
                    <td className="px-6 py-3 font-medium text-gray-300 text-xs font-mono">{log.tool_name}</td>
                    <td className="px-6 py-3">
                      <span
                        className={`px-2 py-0.5 rounded-full text-[10px] font-medium ${
                          log.response_status === "SUCCESS"
                            ? "bg-emerald-500/10 text-emerald-400"
                            : log.response_status === "ERROR"
                              ? "bg-red-500/10 text-red-400"
                              : "bg-gray-800 text-gray-400"
                        }`}
                      >
                        {log.response_status || "—"}
                      </span>
                    </td>
                    <td className="px-6 py-3 text-gray-600 text-xs max-w-xs truncate">
                      {log.response_error || "—"}
                    </td>
                  </tr>
                ))}
                {auditLogs.length === 0 && (
                  <tr>
                    <td colSpan={4} className="px-6 py-12 text-center text-gray-600 text-sm">
                      No audit logs found.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
