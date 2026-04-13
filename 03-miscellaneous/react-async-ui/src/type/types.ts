export interface HackerNewsItem {
    id: number;
    title: string;
    url?: string;
    score: number;
    by: string;
    time: number;
}

export interface HackerNewsSearchHit {
    objectID: string;
    title: string | null;
    story_title: string | null;
    url: string | null;
    story_url: string | null;
    author: string;
    points: number | null;
    created_at_i: number;
}

export interface HackerNewsSearchResponse {
    hits: HackerNewsSearchHit[];
}
