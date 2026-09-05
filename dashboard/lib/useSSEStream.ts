"use client";
import { useEffect, useRef, useState } from "react";

export interface AuditEvent {
  id: number;
  session_id: string;
  event_type: string;
  actor: string;
  reasoning: string;
  outcome: string;
  payload?: any;
  created_at: string;
}

export function useSSEStream(sessionId: string) {
  const [events, setEvents] = useState<AuditEvent[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    // Establish a long-lived connection to the Go backend's Redis Pub/Sub bridge
    const url = `${process.env.NEXT_PUBLIC_MERCHANT_URL}/stream/${sessionId}`;
    const es = new EventSource(url);
    esRef.current = es;

    es.onopen = () => setIsConnected(true);

    es.addEventListener("audit_log", (e) => {
      const parsed: AuditEvent = JSON.parse(e.data);
      setEvents((prev) => {
        // Deduplicate events to handle transient network reconnects gracefully
        const exists = prev.find((ev) => ev.id === parsed.id);
        return exists ? prev : [...prev, parsed];
      });
    });

    es.onerror = () => setIsConnected(false);

    return () => {
      es.close();
      setIsConnected(false);
    };
  }, [sessionId]);

  return { events, isConnected };
}
