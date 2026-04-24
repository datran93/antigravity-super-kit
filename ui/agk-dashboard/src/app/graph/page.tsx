"use client";

import { useState, useEffect, useMemo, useRef, useCallback } from "react";
import { fetchDocReferences, fetchKnowledgeItems } from "../actions";
import { DocReference, KnowledgeItem } from "@/lib/db";
import { Network, ZoomIn, ZoomOut, RotateCcw } from "lucide-react";

interface GraphNode {
  id: string;
  type: string;
  label: string;
  x: number;
  y: number;
  vx: number;
  vy: number;
}

interface GraphEdge {
  source: string;
  target: string;
  relation: string;
}

const TYPE_COLORS: Record<string, string> = {
  task: "#3b82f6",
  ki: "#a855f7",
  anchor: "#f59e0b",
  doc: "#10b981",
  unknown: "#6b7280",
};

export default function GraphPage() {
  const [refs, setRefs] = useState<DocReference[]>([]);
  const [kis, setKis] = useState<KnowledgeItem[]>([]);
  const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const nodesRef = useRef<GraphNode[]>([]);
  const edgesRef = useRef<GraphEdge[]>([]);
  const scaleRef = useRef(1);
  const offsetRef = useRef({ x: 0, y: 0 });
  const draggingRef = useRef<{ node: GraphNode | null; startX: number; startY: number }>({
    node: null,
    startX: 0,
    startY: 0,
  });
  const animRef = useRef<number>(0);

  useEffect(() => {
    const load = async () => {
      const [r, k] = await Promise.all([fetchDocReferences(), fetchKnowledgeItems()]);
      setRefs(r);
      setKis(k);
    };
    load();
  }, []);

  // Build graph from refs
  const { nodes, edges } = useMemo(() => {
    const nodeMap = new Map<string, GraphNode>();
    const edgeList: GraphEdge[] = [];

    const getOrCreate = (type: string, id: string) => {
      const key = `${type}:${id}`;
      if (!nodeMap.has(key)) {
        const angle = Math.random() * Math.PI * 2;
        const radius = 100 + Math.random() * 200;
        nodeMap.set(key, {
          id: key,
          type,
          label: id.length > 25 ? id.slice(0, 22) + "…" : id,
          x: 400 + Math.cos(angle) * radius,
          y: 300 + Math.sin(angle) * radius,
          vx: 0,
          vy: 0,
        });
      }
      return nodeMap.get(key)!;
    };

    // Add KIs as nodes
    kis.forEach((ki) => {
      getOrCreate("ki", ki.ki_path.split("/").pop() || ki.tactic_name);
    });

    // Add references as edges
    refs.forEach((ref) => {
      getOrCreate(ref.source_type, ref.source_id);
      getOrCreate(ref.target_type, ref.target_id);
      edgeList.push({
        source: `${ref.source_type}:${ref.source_id}`,
        target: `${ref.target_type}:${ref.target_id}`,
        relation: ref.relation,
      });
    });

    return { nodes: Array.from(nodeMap.values()), edges: edgeList };
  }, [refs, kis]);

  // Store in refs for animation loop
  useEffect(() => {
    nodesRef.current = nodes.map((n) => ({ ...n }));
    edgesRef.current = edges;
  }, [nodes, edges]);

  // Force-directed simulation + render loop
  const draw = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const ns = nodesRef.current;
    const es = edgesRef.current;
    const W = canvas.width;
    const H = canvas.height;

    // Simple force simulation step
    for (let i = 0; i < ns.length; i++) {
      ns[i].vx *= 0.9;
      ns[i].vy *= 0.9;

      // Repulsion between all nodes
      for (let j = i + 1; j < ns.length; j++) {
        const dx = ns[j].x - ns[i].x;
        const dy = ns[j].y - ns[i].y;
        const dist = Math.max(Math.sqrt(dx * dx + dy * dy), 1);
        const force = 2000 / (dist * dist);
        ns[i].vx -= (dx / dist) * force;
        ns[i].vy -= (dy / dist) * force;
        ns[j].vx += (dx / dist) * force;
        ns[j].vy += (dy / dist) * force;
      }

      // Center gravity
      ns[i].vx += (W / 2 - ns[i].x) * 0.001;
      ns[i].vy += (H / 2 - ns[i].y) * 0.001;
    }

    // Spring forces along edges
    es.forEach((e) => {
      const s = ns.find((n) => n.id === e.source);
      const t = ns.find((n) => n.id === e.target);
      if (!s || !t) return;
      const dx = t.x - s.x;
      const dy = t.y - s.y;
      const dist = Math.max(Math.sqrt(dx * dx + dy * dy), 1);
      const force = (dist - 150) * 0.01;
      s.vx += (dx / dist) * force;
      s.vy += (dy / dist) * force;
      t.vx -= (dx / dist) * force;
      t.vy -= (dy / dist) * force;
    });

    // Apply velocities
    ns.forEach((n) => {
      if (draggingRef.current.node?.id !== n.id) {
        n.x += n.vx;
        n.y += n.vy;
      }
    });

    // Clear and draw
    ctx.clearRect(0, 0, W, H);
    ctx.save();
    ctx.translate(offsetRef.current.x, offsetRef.current.y);
    ctx.scale(scaleRef.current, scaleRef.current);

    // Draw edges
    es.forEach((e) => {
      const s = ns.find((n) => n.id === e.source);
      const t = ns.find((n) => n.id === e.target);
      if (!s || !t) return;
      ctx.beginPath();
      ctx.moveTo(s.x, s.y);
      ctx.lineTo(t.x, t.y);
      ctx.strokeStyle = "rgba(75, 85, 99, 0.4)";
      ctx.lineWidth = 1;
      ctx.stroke();

      // Relation label
      const mx = (s.x + t.x) / 2;
      const my = (s.y + t.y) / 2;
      ctx.fillStyle = "rgba(107, 114, 128, 0.6)";
      ctx.font = "9px sans-serif";
      ctx.textAlign = "center";
      ctx.fillText(e.relation, mx, my - 4);
    });

    // Draw nodes
    ns.forEach((n) => {
      const color = TYPE_COLORS[n.type] || TYPE_COLORS.unknown;
      const isSelected = selectedNode?.id === n.id;
      const radius = isSelected ? 10 : 7;

      // Glow
      if (isSelected) {
        ctx.beginPath();
        ctx.arc(n.x, n.y, 18, 0, Math.PI * 2);
        ctx.fillStyle = color + "20";
        ctx.fill();
      }

      ctx.beginPath();
      ctx.arc(n.x, n.y, radius, 0, Math.PI * 2);
      ctx.fillStyle = color;
      ctx.fill();
      ctx.strokeStyle = isSelected ? "#ffffff" : color + "80";
      ctx.lineWidth = isSelected ? 2 : 1;
      ctx.stroke();

      // Label
      ctx.fillStyle = "#d1d5db";
      ctx.font = "11px sans-serif";
      ctx.textAlign = "center";
      ctx.fillText(n.label, n.x, n.y + radius + 14);
    });

    ctx.restore();
    animRef.current = requestAnimationFrame(draw);
  }, [selectedNode]);

  // Start/stop animation
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const resize = () => {
      canvas.width = canvas.parentElement?.clientWidth || 800;
      canvas.height = canvas.parentElement?.clientHeight || 600;
    };
    resize();
    window.addEventListener("resize", resize);

    animRef.current = requestAnimationFrame(draw);
    return () => {
      cancelAnimationFrame(animRef.current);
      window.removeEventListener("resize", resize);
    };
  }, [draw]);

  // Mouse interaction
  const handleMouseDown = (e: React.MouseEvent) => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    const mx = (e.clientX - rect.left - offsetRef.current.x) / scaleRef.current;
    const my = (e.clientY - rect.top - offsetRef.current.y) / scaleRef.current;

    const hit = nodesRef.current.find((n) => {
      const dx = n.x - mx;
      const dy = n.y - my;
      return dx * dx + dy * dy < 200;
    });

    if (hit) {
      draggingRef.current = { node: hit, startX: mx, startY: my };
      setSelectedNode(hit);
    } else {
      setSelectedNode(null);
    }
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!draggingRef.current.node || !canvasRef.current) return;
    const rect = canvasRef.current.getBoundingClientRect();
    const mx = (e.clientX - rect.left - offsetRef.current.x) / scaleRef.current;
    const my = (e.clientY - rect.top - offsetRef.current.y) / scaleRef.current;
    draggingRef.current.node.x = mx;
    draggingRef.current.node.y = my;
    draggingRef.current.node.vx = 0;
    draggingRef.current.node.vy = 0;
  };

  const handleMouseUp = () => {
    draggingRef.current = { node: null, startX: 0, startY: 0 };
  };

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-6 pt-5 pb-3 border-b border-gray-800/50 shrink-0">
        <h1 className="text-lg font-bold flex items-center gap-2.5">
          <Network size={20} className="text-purple-400" />
          Knowledge Graph
        </h1>
        <div className="flex items-center gap-2">
          <button
            onClick={() => {
              scaleRef.current = Math.min(scaleRef.current * 1.2, 3);
            }}
            className="p-1.5 rounded-md bg-gray-900 hover:bg-gray-800 text-gray-500 transition-colors border border-gray-800"
          >
            <ZoomIn size={14} />
          </button>
          <button
            onClick={() => {
              scaleRef.current = Math.max(scaleRef.current / 1.2, 0.3);
            }}
            className="p-1.5 rounded-md bg-gray-900 hover:bg-gray-800 text-gray-500 transition-colors border border-gray-800"
          >
            <ZoomOut size={14} />
          </button>
          <button
            onClick={() => {
              scaleRef.current = 1;
              offsetRef.current = { x: 0, y: 0 };
            }}
            className="p-1.5 rounded-md bg-gray-900 hover:bg-gray-800 text-gray-500 transition-colors border border-gray-800"
          >
            <RotateCcw size={14} />
          </button>
          <div className="flex items-center gap-3 ml-3 text-[10px] text-gray-600">
            {Object.entries(TYPE_COLORS)
              .filter(([k]) => k !== "unknown")
              .map(([type, color]) => (
                <span key={type} className="flex items-center gap-1">
                  <span className="w-2 h-2 rounded-full inline-block" style={{ backgroundColor: color }} />
                  {type}
                </span>
              ))}
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-hidden relative m-4 bg-gray-900 rounded-xl border border-gray-800">
        <canvas
          ref={canvasRef}
          className="w-full h-full cursor-grab active:cursor-grabbing"
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
        />

        {/* Selected node detail */}
        {selectedNode && (
          <div className="absolute bottom-4 left-4 bg-gray-950/90 backdrop-blur border border-gray-700 rounded-lg p-4 max-w-xs">
            <div className="flex items-center gap-2 mb-2">
              <span
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: TYPE_COLORS[selectedNode.type] || TYPE_COLORS.unknown }}
              />
              <span className="text-xs font-medium text-gray-400 uppercase">{selectedNode.type}</span>
            </div>
            <p className="text-sm font-medium text-gray-200">{selectedNode.label}</p>
            <p className="text-xs text-gray-500 mt-1 font-mono">{selectedNode.id}</p>
          </div>
        )}

        {nodes.length === 0 && (
          <div className="absolute inset-0 flex items-center justify-center text-gray-500">
            <div className="text-center">
              <Network size={48} className="mx-auto mb-3 opacity-30" />
              <p>No references found. Create @doc, @ki, or @task references to see the graph.</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
