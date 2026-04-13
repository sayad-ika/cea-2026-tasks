import type {
    HackerNewsItem,
    HackerNewsSearchHit,
    HackerNewsSearchResponse,
} from "../type/types";

const BASE_URL = "https://hacker-news.firebaseio.com/v0";
const ALGOLIA_BASE_URL = "https://hn.algolia.com/api/v1";

export class ApiError extends Error {
    constructor(message: string) {
        super(message);
        this.name = "ApiError";
    }
}

export async function fetchTopStoryIds(
    signal?: AbortSignal,
): Promise<number[]> {
    const response = await fetch(`${BASE_URL}/newstories.json?print=pretty`, {
        signal,
    });
    if (!response.ok) {
        throw new ApiError(`Failed to fetch story IDs: ${response.status}`);
    }
    const data = await response.json();
    if (!Array.isArray(data)) {
        throw new ApiError("Invalid response format");
    }
    return data;
}

export async function fetchItems(
    ids: number[],
    signal?: AbortSignal,
): Promise<HackerNewsItem[]> {
    try {
        const results = await Promise.all(
            ids.map(async (id) => {
                const response = await fetch(`${BASE_URL}/item/${id}.json`, {
                    signal,
                });

                if (!response.ok) {
                    throw new ApiError(
                        `Failed to fetch item ${id}: ${response.status}`,
                    );
                }

                const item = (await response.json()) as HackerNewsItem | null;
                if (!item) {
                    throw new ApiError(`Empty response for item ${id}`);
                }

                return item;
            }),
        );

        return results;
    } catch (error: unknown) {
        if (error instanceof Error && error.name === "AbortError") {
            throw error;
        }

        if (error instanceof ApiError) {
            throw error;
        }

        throw new ApiError("Failed to fetch story items");
    }
}

function mapHitToItem(hit: HackerNewsSearchHit): HackerNewsItem | null {
    const title = hit.title || hit.story_title;
    if (!title) return null;
    return {
        id: Number(hit.objectID),
        title,
        url: hit.url || hit.story_url || undefined,
        score: hit.points ?? 0,
        by: hit.author,
        time: hit.created_at_i,
    };
}

export async function fetchStories(
    query: string,
    limit: number,
    signal?: AbortSignal,
): Promise<HackerNewsItem[]> {
    const trimmed = query.trim();
    const endpoint = trimmed
        ? `${ALGOLIA_BASE_URL}/search?query=${encodeURIComponent(trimmed)}&tags=story&hitsPerPage=${limit}`
        : `${ALGOLIA_BASE_URL}/search_by_date?tags=story&hitsPerPage=${limit}`;

    const response = await fetch(endpoint, { signal });
    if (!response.ok)
        throw new ApiError(`Failed to fetch stories: ${response.status}`);

    const data = (await response.json()) as HackerNewsSearchResponse;
    return data.hits
        .map(mapHitToItem)
        .filter((item): item is HackerNewsItem => item !== null);
}
