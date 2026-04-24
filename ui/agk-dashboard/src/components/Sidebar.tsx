"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { CheckCircle2, Activity, Network, MessageSquare } from "lucide-react";

const navItems = [
  { href: "/", label: "Tasks", icon: CheckCircle2, color: "blue" },
  { href: "/activity", label: "Activity", icon: Activity, color: "emerald" },
  { href: "/graph", label: "Graph", icon: Network, color: "purple" },
  { href: "/chat", label: "Chat", icon: MessageSquare, color: "amber" },
];

const colorMap: Record<string, { active: string; hover: string }> = {
  blue: { active: "bg-blue-600/20 text-blue-400", hover: "hover:bg-gray-800 text-gray-400" },
  emerald: { active: "bg-emerald-600/20 text-emerald-400", hover: "hover:bg-gray-800 text-gray-400" },
  purple: { active: "bg-purple-600/20 text-purple-400", hover: "hover:bg-gray-800 text-gray-400" },
  amber: { active: "bg-amber-600/20 text-amber-400", hover: "hover:bg-gray-800 text-gray-400" },
};

export default function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-56 bg-gray-900 border-r border-gray-800 flex flex-col shrink-0">
      <div className="p-4 border-b border-gray-800">
        <h1 className="text-lg font-bold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
          AGK Dashboard
        </h1>
        <p className="text-xs text-gray-500 mt-1">Antigravity Kit</p>
      </div>
      <nav className="flex-1 p-3 space-y-1">
        {navItems.map(({ href, label, icon: Icon, color }) => {
          const isActive = pathname === href;
          const colors = colorMap[color];
          return (
            <Link
              key={href}
              href={href}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                isActive ? colors.active : colors.hover
              }`}
            >
              <Icon size={16} />
              {label}
            </Link>
          );
        })}
      </nav>
      <div className="p-4 border-t border-gray-800 text-xs text-gray-600">
        v0.1.0
      </div>
    </aside>
  );
}
