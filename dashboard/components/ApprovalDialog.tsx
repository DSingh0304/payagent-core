export default function ApprovalDialog({ cart, onDecide }: { cart: any; onDecide: (d: "approved" | "rejected") => void }) {
  return (
    <div style={{
      background: "var(--surface-2)",
      border: "1px solid var(--yellow)",
      borderRadius: "var(--radius)",
      padding: 20,
      boxShadow: "0 0 24px rgba(234, 179, 8, 0.1)",
    }}>
      <div style={{ fontSize: 13, fontWeight: 600, color: "var(--yellow)", marginBottom: 8 }}>
        ⚠ Agent Requesting Approval
      </div>
      <p style={{ fontSize: 13, color: "var(--text)", marginBottom: 16, lineHeight: 1.6 }}>
        The agent wants to create a Razorpay payment order for{" "}
        <strong>₹{cart?.total_inr?.toFixed(2)}</strong>. Do you approve?
      </p>
      <div style={{ display: "flex", gap: 10 }}>
        <button
          onClick={() => onDecide("approved")}
          style={{
            flex: 1, padding: "10px 0", borderRadius: 8, border: "none", cursor: "pointer",
            background: "var(--green)", color: "#000", fontWeight: 700, fontSize: 13,
            transition: "opacity 0.2s",
          }}
          onMouseOver={(e) => (e.currentTarget.style.opacity = "0.85")}
          onMouseOut={(e) => (e.currentTarget.style.opacity = "1")}
        >
          Approve
        </button>
        <button
          onClick={() => onDecide("rejected")}
          style={{
            flex: 1, padding: "10px 0", borderRadius: 8, border: "1px solid var(--border)", cursor: "pointer",
            background: "transparent", color: "var(--text)", fontWeight: 600, fontSize: 13,
            transition: "border-color 0.2s",
          }}
          onMouseOver={(e) => (e.currentTarget.style.borderColor = "var(--red)")}
          onMouseOut={(e) => (e.currentTarget.style.borderColor = "var(--border)")}
        >
          Reject
        </button>
      </div>
    </div>
  );
}
