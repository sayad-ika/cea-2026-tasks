import type { HackerNewsItem } from "../type/types";

const BASE_URL = "https://hacker-news.firebaseio.com/v0";

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
                const response = await fetch(
                    `${BASE_URL}/item/${id}.json`,
                    {
                        signal,
                    },
                );

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
