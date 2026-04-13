import { afterEach, describe, expect, it, vi } from "vitest";
import { getRelativeTime } from "./utils";

afterEach(() => {
    vi.useRealTimers();
});

describe("getRelativeTime", () => {
    it("returns unknown when timestamp is missing", () => {
        expect(getRelativeTime(undefined)).toBe("Unknown");
    });

    it("formats minutes", () => {
        vi.useFakeTimers();
        vi.setSystemTime(new Date("2026-04-13T10:00:00Z"));

        const timestamp = Math.floor(
            new Date("2026-04-13T09:55:00Z").getTime() / 1000,
        );
        expect(getRelativeTime(timestamp)).toBe("5 min ago");
    });

    it("formats hours and days", () => {
        vi.useFakeTimers();
        vi.setSystemTime(new Date("2026-04-13T10:00:00Z"));

        const twoHoursAgo = Math.floor(
            new Date("2026-04-13T08:00:00Z").getTime() / 1000,
        );
        const twoDaysAgo = Math.floor(
            new Date("2026-04-11T10:00:00Z").getTime() / 1000,
        );

        expect(getRelativeTime(twoHoursAgo)).toBe("2 hr ago");
        expect(getRelativeTime(twoDaysAgo)).toBe("2 days ago");
    });
});
