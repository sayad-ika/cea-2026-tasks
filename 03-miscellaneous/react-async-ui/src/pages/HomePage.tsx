import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchTopStoryIds, fetchItems } from "../api/hackerNewsApi";
import type { HackerNewsItem } from "@/type/types";
import { getRelativeTime } from "@/util/utils";
import { useQuery } from "@tanstack/react-query";

let cachedPosts: HackerNewsItem[] | null = null;

export default function HomePage() {
    const [hackerItemData, setHackerItemData] = useState<HackerNewsItem[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<boolean>(false);
    const [fetchMethod, setFetchMethod] = useState<"api" | "react-query">(
        "react-query",
    );
    const navigate = useNavigate();

    useEffect(() => {
        if (fetchMethod !== "api") {
            return;
        }

        const controller = new AbortController();
        const signal = controller.signal;

        const fetchDataByFetchApi = async () => {
            setLoading(true);
            if (cachedPosts) {
                setHackerItemData(cachedPosts);
                setLoading(false);
                return;
            }
            try {
                const ids = await fetchTopStoryIds(signal);
                const result = await fetchItems(ids.slice(0, 20), signal);
                cachedPosts = result;
                setHackerItemData(result);
            } catch (error: unknown) {
                if (error instanceof Error && error.name !== "AbortError") {
                    console.error("Error fetching data:", error);
                    setError(true);
                }
            } finally {
                if (!signal.aborted) {
                    setLoading(false);
                }
            }
        };

        fetchDataByFetchApi();

        return () => {
            controller.abort();
        };
    }, [fetchMethod]);

    const {
        data: ids,
        isLoading,
        isError: idsError,
    } = useQuery({
        queryKey: ["topStoryIds"],
        queryFn: ({ signal }) => fetchTopStoryIds(signal),
        staleTime: 1000 * 60 * 5,
        enabled: fetchMethod === "react-query",
    });

    const {
        data: items,
        isLoading: itemsLoading,
        isError: itemsError,
    } = useQuery({
        queryKey: [ids],
        queryFn: ({ signal }) => fetchItems(ids?.slice(0, 20) || [], signal),
        staleTime: 1000 * 60 * 5,
        enabled: !!ids && fetchMethod === "react-query",
    });

    const data = fetchMethod === "react-query" ? items : hackerItemData;
    const isLoadingFinal =
        fetchMethod === "react-query" ? isLoading || itemsLoading : loading;
    const errorFinal =
        fetchMethod === "react-query" ? idsError || itemsError : error;

    return (
        <>
            <h1 className="text-center text-4xl font-bold py-6">
                React Async UI
            </h1>

            <div className="xl:w-2/3 m-auto px-6 space-y-4">
                <div className="flex justify-between items-center">
                    <h2 className="text-xl font-bold">Hacker News Posts</h2>

                    <div className="flex gap-2">
                        <button
                            onClick={() => setFetchMethod("api")}
                            className="px-3 py-1.5 text-sm font-medium border rounded-md bg-white hover:bg-gray-100 transition"
                        >
                            Fetch API
                        </button>

                        <button
                            onClick={() => setFetchMethod("react-query")}
                            className="px-3 py-1.5 text-sm font-medium border rounded-md bg-white hover:bg-gray-100 transition"
                        >
                            React Query
                        </button>
                    </div>
                </div>

                {isLoadingFinal && !errorFinal && (
                    <div className="space-y-2">
                        {Array.from({ length: 20 }).map((_, i) => (
                            <div
                                key={i}
                                className="h-10 w-full animate-pulse bg-gray-200 rounded"
                            />
                        ))}
                    </div>
                )}

                {!isLoadingFinal && errorFinal && (
                    <div className="text-center space-y-3">
                        <div className="bg-red-200 text-red-white p-4 rounded">
                            Something went wrong
                        </div>

                        <button
                            onClick={() => location.reload()}
                            className="px-3 py-1.5 text-sm font-medium rounded-md bg-black text-white hover:cursor-pointer transition"
                        >
                            Retry
                        </button>
                    </div>
                )}

                {!isLoadingFinal && !errorFinal && data?.length === 0 && (
                    <div className="w-full flex justify-center items-center py-20">
                        <div className="flex flex-col items-center text-center gap-4 max-w-sm">
                            <div className="bg-gray-100 p-4 rounded-full">
                                <svg
                                    className="w-6 h-6 text-gray-500"
                                    fill="none"
                                    stroke="currentColor"
                                    strokeWidth="2"
                                    viewBox="0 0 24 24"
                                >
                                    <path
                                        strokeLinecap="round"
                                        strokeLinejoin="round"
                                        d="M9 12h6m-6 4h6M7 8h10M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                                    />
                                </svg>
                            </div>

                            <h2 className="text-xl font-semibold">
                                No stories found
                            </h2>

                            <p className="text-sm text-gray-500">
                                Nothing to show right now. Try refreshing the
                                page.
                            </p>

                            <button
                                onClick={() => window.location.reload()}
                                className="px-3 py-1.5 text-sm font-medium border rounded-md bg-white hover:bg-gray-100 transition"
                            >
                                Refresh
                            </button>
                        </div>
                    </div>
                )}

                {!isLoadingFinal && !errorFinal && data && data?.length > 0 && (
                    <div>
                        <table className="w-full border-collapse">
                            <thead>
                                <tr className="border-b text-left">
                                    <th className="p-2">Story</th>
                                    <th className="p-2">Score</th>
                                    <th className="p-2">Author</th>
                                    <th className="p-2 text-right">Time</th>
                                </tr>
                            </thead>

                            <tbody>
                                {data?.map((item) => (
                                    <tr
                                        key={item.id}
                                        onClick={() =>
                                            navigate(`/post/${item.id}`)
                                        }
                                        className="border-b hover:bg-gray-50 cursor-pointer transition"
                                    >
                                        <td className="p-2">
                                            <div className="font-medium truncate">
                                                {item.title}
                                            </div>

                                            <div className="text-xs text-gray-500">
                                                {item.url
                                                    ? new URL(
                                                          item.url,
                                                      ).hostname.replace(
                                                          "www.",
                                                          "",
                                                      )
                                                    : "news.ycombinator.com"}
                                            </div>
                                        </td>

                                        <td className="p-2 font-medium">
                                            {item.score}
                                        </td>

                                        <td className="p-2 text-gray-500">
                                            {item.by}
                                        </td>

                                        <td className="p-2 text-right text-gray-500">
                                            {getRelativeTime(item.time)}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>

                        <footer className="w-full text-center text-gray-500 text-sm py-4">
                            Top Hacker News stories
                        </footer>
                    </div>
                )}
            </div>
        </>
    );
}
