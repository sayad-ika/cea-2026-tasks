import { useEffect, useRef, useState } from "react";
import type { HeadcountData } from "../types";

export function useHeadcountStream(date: string, isAuthenticated: boolean) {
  const [data, setData] = useState<HeadcountData | null>(null);
  const [streamError, setStreamError] = useState<string | null>(null);
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!date || !isAuthenticated) return;

    const url = `${import.meta.env.VITE_API_BASE_URL}/headcount/${date}/stream`;
    const es = new EventSource(url, { withCredentials: true });
    esRef.current = es;

    es.onmessage = (event) => {
      try {
        setData(JSON.parse(event.data));
        setStreamError(null);
      } catch {
        setStreamError("Failed to parse live data.");
      }
    };

    es.onerror = () => {
      setStreamError("Connection lost — reconnecting...");
    };

    return () => {
      es.close();
      esRef.current = null;
    };
  }, [date, isAuthenticated]);

  return { data, streamError };
}
