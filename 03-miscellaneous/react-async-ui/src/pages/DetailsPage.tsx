import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";

interface HackerNewsItem {
    id: number;
    title: string;
    url: string;
    score: number;
    by: string;
    time: number;
}

export default function DetailsPage() {
    const { id } = useParams<{ id: string }>();
    const [selectedItem, setSelectedItem] = useState<HackerNewsItem | null>(
        null,
    );
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<boolean>(false);
    const navigate = useNavigate();

    useEffect(() => {
        const controller = new AbortController();
        const signal = controller.signal;

        const fetchData = async () => {
            if (!id) return;

            setLoading(true);

            try {
                const response = await fetch(
                    `https://hacker-news.firebaseio.com/v0/item/${id}.json?print=pretty`,
                    { signal },
                );

                if (!response.ok) {
                    throw new Error("Invalid response");
                }

                const itemData = await response.json();
                setSelectedItem(itemData);
            } catch (error: unknown) {
                if (
                    error instanceof Error &&
                    error?.name !== "AbortError"
                ) {
                    console.error("Error fetching item data:", error);
                    setError(true);
                }
            } finally {
                if (!signal.aborted) {
                    setLoading(false);
                }
            }
        };

        fetchData();

        return () => {
            controller.abort();
        };
    }, [id]);

    const getRelativeTime = (timestamp: number | undefined) => {
        if (!timestamp) return "Unknown";
        const diff = Date.now() - timestamp * 1000;
        const minutes = Math.floor(diff / 60000);

        if (minutes < 60) return `${minutes} min ago`;
        const hours = Math.floor(minutes / 60);
        if (hours < 24) return `${hours} hr ago`;
        const days = Math.floor(hours / 24);
        return `${days} days ago`;
    };

    if (loading) {
        return (
            <div className="xl:w-2/3 m-auto p-10 space-y-4">
                <div className="h-8 w-24 animate-pulse bg-gray-200 rounded" />

                <div className="border rounded-xl p-6 space-y-4 shadow">
                    <div className="h-6 w-3/4 animate-pulse bg-gray-200 rounded" />
                    <div className="h-4 w-1/2 animate-pulse bg-gray-200 rounded" />
                </div>
            </div>
        );
    }

    if (error || !selectedItem) {
        return (
            <div className="xl:w-2/3 m-auto p-10 space-y-4">
                <button
                    onClick={() => navigate("/")}
                    className="px-3 py-1.5 text-sm font-medium border rounded-md bg-white hover:bg-gray-100 transition"
                >
                    ← Back
                </button>

                <div className="bg-red-200 text-red-white p-4 rounded">
                    Something went wrong
                </div>
            </div>
        );
    }

    return (
        <div className="xl:w-2/3 m-auto p-10 space-y-4">
            <button
                onClick={() => navigate("/")}
                className="px-3 py-1.5 text-sm font-medium border rounded-md bg-white hover:bg-gray-100 transition"
            >
                ← Back
            </button>

            <div className="border rounded-xl p-6 shadow space-y-3">
                <h2 className="text-2xl font-semibold">
                    {selectedItem?.title}
                </h2>

                <div className="text-sm text-gray-500 flex gap-4">
                    <span>👤 {selectedItem?.by}</span>
                    <span>⭐ {selectedItem?.score}</span>
                    <span>🕒 {getRelativeTime(selectedItem?.time)}</span>
                </div>

                <a
                    href={
                        selectedItem?.url ||
                        `https://news.ycombinator.com/item?id=${selectedItem?.id}`
                    }
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 underline break-all"
                >
                    {selectedItem?.url || "Open on Hacker News"}
                </a>
            </div>
        </div>
    );
}
