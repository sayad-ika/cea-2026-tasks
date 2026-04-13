import { beforeEach, describe, expect, it, vi } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import DetailsPage from "./DetailsPage";
import { renderWithProviders } from "@/test/test-utils";
import type { HackerNewsItem } from "@/type/types";

const { navigateMock } = vi.hoisted(() => ({
    navigateMock: vi.fn(),
}));

vi.mock("react-router-dom", async () => {
    const actual = await vi.importActual<typeof import("react-router-dom")>(
        "react-router-dom",
    );

    return {
        ...actual,
        useNavigate: () => navigateMock,
    };
});

vi.mock("../api/hackerNewsApi", () => ({
    fetchItems: vi.fn(),
}));

import { fetchItems } from "../api/hackerNewsApi";

const fetchItemsMock = vi.mocked(fetchItems);

const STORY: HackerNewsItem = {
    id: 1,
    title: "Story details",
    url: "https://news.ycombinator.com/item?id=1",
    score: 55,
    by: "pg",
    time: 1_710_000_000,
};

const STORY_WITHOUT_URL: HackerNewsItem = {
    ...STORY,
    id: 2,
    url: undefined,
};

describe("DetailsPage", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        navigateMock.mockReset();
    });

    it("renders story detail on success", async () => {
        fetchItemsMock.mockResolvedValueOnce([STORY]);

        renderWithProviders(<DetailsPage />, {
            route: "/post/1",
            path: "/post/:id",
        });

        expect(await screen.findByText("Story details")).toBeInTheDocument();
        // "pg" is rendered as "👤 pg" inside a span — use regex to match partial text
        expect(screen.getByText(/pg/)).toBeInTheDocument();
        expect(screen.getByText(/55/)).toBeInTheDocument();
    });

    it("shows error state on fetch failure", async () => {
        fetchItemsMock.mockRejectedValueOnce(new Error("Request failed"));

        renderWithProviders(<DetailsPage />, {
            route: "/post/1",
            path: "/post/:id",
        });

        expect(await screen.findByText("Something went wrong")).toBeInTheDocument();
        // Verify retry button is present (retry calls location.reload which is not supported in jsdom)
        expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
    });

    it("falls back to Hacker News link and navigates back", async () => {
        fetchItemsMock.mockResolvedValueOnce([STORY_WITHOUT_URL]);

        renderWithProviders(<DetailsPage />, {
            route: "/post/2",
            path: "/post/:id",
        });

        const user = userEvent.setup();

        const link = await screen.findByRole("link", {
            name: "Open on Hacker News",
        });

        expect(link).toHaveAttribute(
            "href",
            "https://news.ycombinator.com/item?id=2",
        );

        await user.click(screen.getByRole("button", { name: /Back/ }));

        expect(navigateMock).toHaveBeenCalledWith("/");
    });
});

