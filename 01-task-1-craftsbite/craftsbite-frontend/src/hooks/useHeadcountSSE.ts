import { useEffect, useRef } from "react";
import type { HeadcountReportDay } from "../types/meal.types";

interface HeadcountSSEPayload {
    success: boolean;
    data: HeadcountReportDay[];
}

/**
 * Subscribes to the SSE stream and calls onUpdate whenever
 * the server broadcasts a new headcount payload.
 *
 * withCredentials: true — browser sends the auth_token cookie automatically.
 * No token handling needed here since we migrated to HttpOnly cookies.
 */
export function useHeadcountSSE(
    onUpdate: (data: HeadcountReportDay[]) => void,
    enabled = true  // pass false for non-admin/logistics roles so they never connect
) {
    const onUpdateRef = useRef(onUpdate);

    // Keep ref in sync without re-running the effect
    useEffect(() => {
        onUpdateRef.current = onUpdate;
    }, [onUpdate]);

    useEffect(() => {
        if (!enabled) return;

        const es = new EventSource(
            `${import.meta.env.VITE_API_BASE_URL}/headcount/report/live`,
            { withCredentials: true } // sends auth_token cookie automatically
        );

        es.addEventListener("headcount-update", (event: MessageEvent) => {
            try {
                const payload = JSON.parse(event.data) as HeadcountSSEPayload;
                if (payload.success && payload.data) {
                    onUpdateRef.current(payload.data);
                }
            } catch (err) {
                console.error("[SSE] Failed to parse headcount update:", err);
            }
        });

        es.onerror = () => {
            // EventSource auto-reconnects after ~3 seconds — nothing to handle manually
            console.warn("[SSE] Connection lost, reconnecting…");
        };

        // Cleanup: close connection when component unmounts or role changes
        return () => {
            es.close();
        };
    }, [enabled]);
}