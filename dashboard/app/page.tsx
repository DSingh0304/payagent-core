"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import SessionStarter from "@/components/SessionStarter";

export default function Home() {
  const router = useRouter();

  const handleStart = async (goal: string) => {
    const res = await fetch(`${process.env.NEXT_PUBLIC_AGENT_URL}/agent/run`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ goal }),
    });
    const data = await res.json();
    router.push(`/session/${data.session_id}`);
  };

  return (
    <main style={{ display: "flex", alignItems: "center", justifyContent: "center", minHeight: "100vh" }}>
      <SessionStarter onStart={handleStart} />
    </main>
  );
}
