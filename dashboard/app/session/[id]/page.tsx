"use client";
import { use, useState, useEffect } from "react";
import AuditTimeline from "@/components/AuditTimeline";
import CartSummary from "@/components/CartSummary";
import ApprovalDialog from "@/components/ApprovalDialog";
import { useSSEStream } from "@/lib/useSSEStream";

export default function SessionPage({ params }: { params: { id: string } }) {
  const { id } = params;
  const { events, isConnected } = useSSEStream(id);
  const [showApproval, setShowApproval] = useState(false);
  const [cart, setCart] = useState<any>(null);

  useEffect(() => {
    const latest = events[events.length - 1];
    if (latest?.event_type === "APPROVAL_REQUIRED") {
      setShowApproval(true);
    }
    if (latest?.event_type === "CART_ADD") {
      fetchCart();
    }
  }, [events]);

  const fetchCart = async () => {
    const res = await fetch(
      `${process.env.NEXT_PUBLIC_MERCHANT_URL}/api/v1/cart/${id}`,
      { headers: { "X-API-Key": "internal-agent-api-key-change-in-prod" } }
    );
    setCart(await res.json());
  };

  const handleDecision = async (decision: "approved" | "rejected") => {
    await fetch(`${process.env.NEXT_PUBLIC_MERCHANT_URL}/agent/${id}/resume`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ decision }),
    });
    setShowApproval(false);
  };

  return (
    <div style={{ maxWidth: 1100, margin: "0 auto", padding: "40px 24px", display: "grid", gridTemplateColumns: "1fr 360px", gap: 24 }}>
      <div>
        <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 28 }}>
          <div style={{
            width: 8, height: 8, borderRadius: "50%",
            background: isConnected ? "var(--green)" : "var(--text-muted)",
            boxShadow: isConnected ? "0 0 8px var(--green)" : "none",
          }} />
          <h1 style={{ fontSize: 18, fontWeight: 600, color: "var(--text)" }}>Agent Session</h1>
          <code style={{ fontSize: 12, color: "var(--text-muted)", background: "var(--surface-2)", padding: "2px 8px", borderRadius: 6 }}>{id}</code>
        </div>
        <AuditTimeline events={events} />
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
        <CartSummary cart={cart} />
        {showApproval && <ApprovalDialog cart={cart} onDecide={handleDecision} />}
      </div>
    </div>
  );
}
