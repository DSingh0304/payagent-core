import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "PayAgent — AI Shopping Dashboard",
  description: "Real-time AI agent audit trail for autonomous commerce",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
