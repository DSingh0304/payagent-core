export default function CartSummary({ cart }: { cart: any }) {
  if (!cart) {
    return (
      <div style={{ background: "var(--surface)", border: "1px solid var(--border)", borderRadius: "var(--radius)", padding: 20 }}>
        <h2 style={{ fontSize: 14, fontWeight: 600, marginBottom: 8, color: "var(--text-muted)" }}>Cart</h2>
        <p style={{ fontSize: 13, color: "var(--text-muted)" }}>Empty</p>
      </div>
    );
  }

  return (
    <div style={{ background: "var(--surface)", border: "1px solid var(--border)", borderRadius: "var(--radius)", padding: 20 }}>
      <h2 style={{ fontSize: 14, fontWeight: 600, marginBottom: 14 }}>Current Cart</h2>
      <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
        {cart.items?.map((item: any) => (
          <div key={item.product_id} style={{ display: "flex", justifyContent: "space-between", fontSize: 13 }}>
            <div>
              <div style={{ fontWeight: 500 }}>{item.name}</div>
              <div style={{ fontSize: 11, color: "var(--text-muted)" }}>qty: {item.quantity}</div>
            </div>
            <div style={{ fontWeight: 600 }}>₹{item.price_inr * item.quantity}</div>
          </div>
        ))}
      </div>
      <div style={{ borderTop: "1px solid var(--border)", marginTop: 14, paddingTop: 12, display: "flex", justifyContent: "space-between", fontWeight: 700 }}>
        <span>Total</span>
        <span style={{ color: "var(--accent)" }}>₹{cart.total_inr?.toFixed(2)}</span>
      </div>
    </div>
  );
}
