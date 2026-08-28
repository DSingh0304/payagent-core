import { AuditEvent } from "@/lib/useSSEStream";

const EVENT_COLORS: Record<string, string> = {
  CART_ADD: "#6c63ff",
  CART_REMOVE: "#ef4444",
  APPROVAL_REQUIRED: "#eab308",
  ORDER_CREATED: "#22c55e",
};

export default function AuditTimeline({ events }: { events: AuditEvent[] }) {
  if (events.length === 0) {
    return (
      <div style={{ textAlign: "center", padding: "60px 0", color: "var(--text-muted)" }}>
        <div style={{ fontSize: 32, marginBottom: 12 }}>⏳</div>
        <p>Waiting for agent activity...</p>
      </div>
    );
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
      {events.map((ev) => (
        <div key={ev.id} style={{
          background: "var(--surface)",
          border: "1px solid var(--border)",
          borderRadius: "var(--radius)",
          padding: "14px 16px",
          borderLeft: `3px solid ${EVENT_COLORS[ev.event_type] || "var(--accent)"}`,
          animation: "fadeIn 0.3s ease",
        }}>
          <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 6 }}>
            <span style={{
              fontSize: 11, fontWeight: 600, letterSpacing: "0.05em",
              color: EVENT_COLORS[ev.event_type] || "var(--accent)",
              textTransform: "uppercase",
            }}>
              {ev.event_type}
            </span>
            <span style={{ fontSize: 11, color: "var(--text-muted)" }}>
              {new Date(ev.created_at).toLocaleTimeString()}
            </span>
          </div>
          {ev.reasoning && (
            <p style={{ fontSize: 13, color: "var(--text)", lineHeight: 1.5 }}>{ev.reasoning}</p>
          )}
          <div style={{ fontSize: 11, color: "var(--text-muted)", marginTop: 4 }}>
            actor: {ev.actor} · outcome: {ev.outcome || "pending"}
          </div>
        </div>
      ))}
      <style>{`@keyframes fadeIn { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: none; } }`}</style>
    </div>
  );
}
