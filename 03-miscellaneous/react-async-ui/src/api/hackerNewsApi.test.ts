import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, fetchStories, fetchItems, fetchTopStoryIds } from "./hackerNewsApi";

function jsonResponse(body: unknown, status = 200): Response {
    return new Response(JSON.stringify(body), {
        status,
        headers: { "Content-Type": "application/json" },
    });
}

afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
});

describe("fetchStories", () => {
    it("maps Algolia response to HackerNewsItem[]", async () => {
        vi.stubGlobal(
            "fetch",
            vi.fn().mockResolvedValue(
                jsonResponse({
                    hits: [
                        {
                            objectID: "123",
                            title: "React article",
                            story_title: null,
                            url: "https://react.dev/blog",
                            story_url: null,
                            author: "dan",
                            points: 10,
                            created_at_i: 1_710_000_000,
                            num_comments: 2,
                            story_text: null,
                            comment_text: null,
                            _tags: ["story"],
                        },
                    ],
                }),
            ),
        );

        const result = await fetchStories("react", 20);

        expect(result).toHaveLength(1);
        expect(result[0]).toMatchObject({
            id: 123,
            title: "React article",
            by: "dan",
        });
    });

    it("throws ApiError on non-200 response", async () => {
        vi.stubGlobal(
            "fetch",
            vi.fn().mockResolvedValue(jsonResponse({}, 500)),
        );

        await expect(fetchStories("react", 20)).rejects.toBeInstanceOf(
            ApiError,
        );
    });

    it("uses search_by_date endpoint when query is empty", async () => {
        const mockFetch = vi.fn().mockResolvedValue(jsonResponse({ hits: [] }));
        vi.stubGlobal("fetch", mockFetch);

        await fetchStories("", 20);

        expect(mockFetch).toHaveBeenCalledWith(
            expect.stringContaining("search_by_date"),
            expect.any(Object),
        );
    });
});

describe("fetchItems", () => {
    it("returns mapped items", async () => {
        vi.stubGlobal(
            "fetch",
            vi.fn().mockResolvedValue(
                jsonResponse({
                    id: 1,
                    title: "Item title",
                    score: 12,
                    by: "pg",
                    time: 1_710_000_000,
                }),
            ),
        );

        const items = await fetchItems([1]);

        expect(items[0]?.id).toBe(1);
        expect(items[0]?.title).toBe("Item title");
    });

    it("throws ApiError on non-200 response", async () => {
        vi.stubGlobal(
            "fetch",
            vi.fn().mockResolvedValue(jsonResponse({}, 404)),
        );

        await expect(fetchItems([1])).rejects.toBeInstanceOf(ApiError);
    });
});

describe("fetchTopStoryIds", () => {
    it("returns array of ids", async () => {
        vi.stubGlobal(
            "fetch",
            vi.fn().mockResolvedValue(jsonResponse([1, 2, 3])),
        );

        const ids = await fetchTopStoryIds();

        expect(ids).toEqual([1, 2, 3]);
    });

    it("throws ApiError on non-200 response", async () => {
        vi.stubGlobal(
            "fetch",
            vi.fn().mockResolvedValue(jsonResponse({}, 500)),
        );

        await expect(fetchTopStoryIds()).rejects.toBeInstanceOf(ApiError);
    });

    it("throws ApiError when response is not an array", async () => {
        vi.stubGlobal(
            "fetch",
            vi.fn().mockResolvedValue(jsonResponse({ invalid: true })),
        );

        await expect(fetchTopStoryIds()).rejects.toBeInstanceOf(ApiError);
    });
});

// add inside existing fetchStories describe:
it("filters out hits with no title", async () => {
    vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue(
            jsonResponse({
                hits: [
                    {
                        objectID: "1",
                        title: null,
                        story_title: null, // both null → filtered out
                        url: null,
                        story_url: null,
                        author: "dan",
                        points: 10,
                        created_at_i: 1_710_000_000,
                        _tags: ["story"],
                    },
                ],
            }),
        ),
    );

    const result = await fetchStories("react", 20);

    expect(result).toHaveLength(0);
});
