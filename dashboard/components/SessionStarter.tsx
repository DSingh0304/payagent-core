"use client";
import { useState } from "react";

export default function SessionStarter({ onStart }: { onStart: (goal: string) => void }) {
  const [goal, setGoal] = useState("");
  const [loading, setLoading] = useState(false);

  const examples = [
    "Buy running shoes under ₹2500",
    "Get a birthday gift for my sister under ₹1000",
    "Find the cheapest laptop bag in stock",
  ];

  const submit = async () => {
    if (!goal.trim()) return;
    setLoading(true);
    await onStart(goal);
  };

  return (
    <div style={{ width: "100%", maxWidth: 520, padding: "0 24px" }}>
      <div style={{ marginBottom: 32, textAlign: "center" }}>
        <h1 style={{ fontSize: 32, fontWeight: 700, marginBottom: 8 }}>PayAgent</h1>
        <p style={{ color: "var(--text-muted)", fontSize: 15 }}>Give the AI a shopping goal and watch it work</p>
      </div>

      <div style={{
        background: "var(--surface)",
        border: "1px solid var(--border)",
        borderRadius: "var(--radius)",
        padding: 24,
      }}>
        <textarea
          value={goal}
          onChange={(e) => setGoal(e.target.value)}
          placeholder="e.g. Buy running shoes under ₹2500"
          rows={3}
          style={{
            width: "100%", background: "var(--surface-2)", border: "1px solid var(--border)",
            borderRadius: 8, padding: "12px 14px", fontSize: 14, color: "var(--text)",
            resize: "none", outline: "none", fontFamily: "inherit",
          }}
          onKeyDown={(e) => { if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); submit(); } }}
        />

        <div style={{ display: "flex", flexWrap: "wrap", gap: 8, margin: "12px 0" }}>
          {examples.map((ex) => (
            <button key={ex} onClick={() => setGoal(ex)} style={{
              fontSize: 11, padding: "4px 10px", borderRadius: 20,
              border: "1px solid var(--border)", background: "transparent",
              color: "var(--text-muted)", cursor: "pointer",
            }}>
              {ex}
            </button>
          ))}
        </div>

        <button
          onClick={submit}
          disabled={loading || !goal.trim()}
          style={{
            width: "100%", padding: "12px 0", borderRadius: 8, border: "none",
            background: "var(--accent)", color: "#fff", fontWeight: 700,
            fontSize: 14, cursor: loading ? "not-allowed" : "pointer",
            opacity: loading ? 0.7 : 1, transition: "opacity 0.2s",
          }}
        >
          {loading ? "Starting agent..." : "Run Agent →"}
        </button>
      </div>
    </div>
  );
}
