"use client";

import { useState, useRef, useEffect } from "react";
import { MessageSquare, Send, Bot, User, Loader2 } from "lucide-react";

interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
  timestamp: Date;
}

export default function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: "welcome",
      role: "assistant",
      content:
        "👋 Welcome to the AGK Chat! I can help answer questions about your project context.\n\n" +
        "**Note:** This panel requires an MCP proxy endpoint to be configured. " +
        "Set the `AGK_CHAT_ENDPOINT` environment variable to your MCP server URL to enable contextual Q&A.\n\n" +
        "In the meantime, you can use this as a local notepad for your development notes.",
      timestamp: new Date(),
    },
  ]);
  const [input, setInput] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = async () => {
    const text = input.trim();
    if (!text || isLoading) return;

    const userMsg: Message = {
      id: `user-${Date.now()}`,
      role: "user",
      content: text,
      timestamp: new Date(),
    };
    setMessages((prev) => [...prev, userMsg]);
    setInput("");
    setIsLoading(true);

    // Attempt to call MCP endpoint if configured
    try {
      const endpoint = process.env.NEXT_PUBLIC_AGK_CHAT_ENDPOINT;
      if (endpoint) {
        const res = await fetch(endpoint, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ query: text }),
        });
        const data = await res.json();
        setMessages((prev) => [
          ...prev,
          {
            id: `bot-${Date.now()}`,
            role: "assistant",
            content: data.response || data.result || JSON.stringify(data),
            timestamp: new Date(),
          },
        ]);
      } else {
        setMessages((prev) => [
          ...prev,
          {
            id: `bot-${Date.now()}`,
            role: "assistant",
            content:
              "⚙️ No chat endpoint configured.\n\n" +
              "To enable contextual Q&A, set `NEXT_PUBLIC_AGK_CHAT_ENDPOINT` in your environment and restart the dashboard.\n\n" +
              "Your message has been recorded locally.",
            timestamp: new Date(),
          },
        ]);
      }
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        {
          id: `bot-${Date.now()}`,
          role: "assistant",
          content: `❌ Failed to reach chat endpoint: ${err instanceof Error ? err.message : "Unknown error"}`,
          timestamp: new Date(),
        },
      ]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-6 pt-5 pb-3 border-b border-gray-800/50 shrink-0">
        <h1 className="text-lg font-bold flex items-center gap-2.5">
          <MessageSquare size={20} className="text-amber-400" />
          Contextual Chat
        </h1>
        <span className="text-[10px] text-gray-600">{messages.length - 1} messages</span>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
        {messages.map((msg) => (
          <div
            key={msg.id}
            className={`flex gap-3 ${msg.role === "user" ? "justify-end" : "justify-start"}`}
          >
            {msg.role === "assistant" && (
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-amber-500/20 to-orange-500/20 border border-amber-500/30 flex items-center justify-center shrink-0 mt-1">
                <Bot size={16} className="text-amber-400" />
              </div>
            )}
            <div
              className={`max-w-[70%] rounded-xl p-4 text-sm whitespace-pre-wrap ${
                msg.role === "user"
                  ? "bg-blue-600/20 border border-blue-500/30 text-blue-100"
                  : "bg-gray-800/80 border border-gray-700 text-gray-300"
              }`}
            >
              {msg.content}
              <div className="text-xs text-gray-600 mt-2">
                {msg.timestamp.toLocaleTimeString()}
              </div>
            </div>
            {msg.role === "user" && (
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500/20 to-indigo-500/20 border border-blue-500/30 flex items-center justify-center shrink-0 mt-1">
                <User size={16} className="text-blue-400" />
              </div>
            )}
          </div>
        ))}
        {isLoading && (
          <div className="flex gap-3">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-amber-500/20 to-orange-500/20 border border-amber-500/30 flex items-center justify-center shrink-0">
              <Loader2 size={16} className="text-amber-400 animate-spin" />
            </div>
            <div className="bg-gray-800/80 border border-gray-700 rounded-xl p-4 text-sm text-gray-500">
              Thinking...
            </div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className="border-t border-gray-800/50 px-6 py-4 shrink-0">
        <div className="bg-gray-900 rounded-xl border border-gray-800 p-3 flex items-end gap-3">
          <textarea
            ref={inputRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask about your project context... (Enter to send, Shift+Enter for newline)"
            rows={1}
            className="flex-1 bg-transparent text-gray-200 placeholder-gray-600 resize-none focus:outline-none text-sm py-2"
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || isLoading}
            className="p-2.5 rounded-lg bg-amber-600/20 text-amber-400 hover:bg-amber-600/30 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            <Send size={16} />
          </button>
        </div>
      </div>
    </div>
  );
}
