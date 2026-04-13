import { beforeEach, describe, expect, it, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import HomePage from "./HomePage";
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
    fetchTopStoryIds: vi.fn(),
    fetchItems: vi.fn(),
    fetchStories: vi.fn(),
}));

import {
    fetchTopStoryIds,
    fetchItems,
    fetchStories,
} from "../api/hackerNewsApi";

const fetchTopStoryIdsMock = vi.mocked(fetchTopStoryIds);
const fetchItemsMock = vi.mocked(fetchItems);
const fetchStoriesMock = vi.mocked(fetchStories);

const STORY: HackerNewsItem = {
    id: 1,
    title: "React 19 released",
    url: "https://react.dev",
    score: 100,
    by: "gaearon",
    time: 1_710_000_000,
};

const SEARCH_STORY: HackerNewsItem = {
    id: 2,
    title: "Search result story",
    url: "https://example.com",
    score: 50,
    by: "someone",
    time: 1_710_000_000,
};

const STORY_WITHOUT_URL: HackerNewsItem = {
    ...STORY,
    id: 3,
    title: "Offline story",
    url: undefined,
};

const SEARCH_STORY_WITHOUT_URL: HackerNewsItem = {
    ...SEARCH_STORY,
    id: 4,
    title: "Search result without url",
    url: undefined,
};

describe("HomePage", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        navigateMock.mockReset();
        fetchTopStoryIdsMock.mockResolvedValue([1]);
        fetchItemsMock.mockResolvedValue([STORY]);
        fetchStoriesMock.mockResolvedValue([]);
    });

    it("shows loading skeletons while request is pending", () => {
        fetchTopStoryIdsMock.mockImplementation(() => new Promise(() => {}));

        const { container } = renderWithProviders(<HomePage />);

        expect(container.querySelector(".animate-pulse")).toBeInTheDocument();
    });

    it("shows error state when react-query fetch fails", async () => {
        fetchTopStoryIdsMock.mockRejectedValueOnce(new Error("Network down"));

        renderWithProviders(<HomePage />);

        expect(
            await screen.findByText("Something went wrong"),
        ).toBeInTheDocument();
    });

    it("shows empty state when API returns no stories", async () => {
        fetchTopStoryIdsMock.mockResolvedValueOnce([]);
        fetchItemsMock.mockResolvedValueOnce([]);

        renderWithProviders(<HomePage />);

        expect(await screen.findByText("No stories found")).toBeInTheDocument();
    });

    it("shows success state with story data", async () => {
        renderWithProviders(<HomePage />);

        expect(
            await screen.findByText("React 19 released"),
        ).toBeInTheDocument();
        expect(screen.getByText("gaearon")).toBeInTheDocument();
    });

    it("calls fetchStories with query when user types in search", async () => {
        fetchStoriesMock.mockResolvedValue([SEARCH_STORY]);

        renderWithProviders(<HomePage />);

        const input = screen.getByRole("searchbox");
        await userEvent.type(input, "react");

        await waitFor(() => {
            expect(fetchStoriesMock).toHaveBeenCalledWith(
                "react",
                20,
                expect.any(AbortSignal),
            );
        });
    });

    it("shows search results when query is active", async () => {
        fetchStoriesMock.mockResolvedValue([SEARCH_STORY]);

        renderWithProviders(<HomePage />);

        const input = screen.getByRole("searchbox");
        await userEvent.type(input, "react");

        expect(
            await screen.findByText("Search result story"),
        ).toBeInTheDocument();
    });

    it("loads stories through fetch api and handles row interactions", async () => {
        fetchItemsMock.mockResolvedValue([STORY_WITHOUT_URL]);

        renderWithProviders(<HomePage />);

        const user = userEvent.setup();

        await user.click(screen.getByRole("button", { name: "Fetch API" }));

        expect(await screen.findByText("Offline story")).toBeInTheDocument();
        expect(screen.getByText("news.ycombinator.com")).toBeInTheDocument();

        await user.click(screen.getByRole("button", { name: "React Query" }));
        await user.click(screen.getByText("Offline story"));

        expect(navigateMock).toHaveBeenCalledWith("/post/3");
    });

    it("shows fallback hostname for search results and navigates on click", async () => {
        fetchStoriesMock.mockResolvedValue([SEARCH_STORY_WITHOUT_URL]);

        renderWithProviders(<HomePage />);

        const user = userEvent.setup();

        await user.type(screen.getByRole("searchbox"), "react");

        expect(
            await screen.findByText("Search result without url"),
        ).toBeInTheDocument();
        expect(screen.getAllByText("news.ycombinator.com")[0]).toBeInTheDocument();

        await user.click(screen.getByText("Search result without url"));

        expect(navigateMock).toHaveBeenCalledWith("/post/4");
    });
});

