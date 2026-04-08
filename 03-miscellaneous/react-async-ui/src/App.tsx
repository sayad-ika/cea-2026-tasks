import { useEffect, useState } from "react";
import "./App.css";
import {
    Table,
    TableBody,
    TableCaption,
    TableCell,
    TableFooter,
    TableHead,
    TableHeader,
    TableRow,
} from "./components/ui/table";
import { Skeleton } from "./components/ui/skeleton";
import { AlertCircle, FileText } from "lucide-react";
import { Button } from "./components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "./components/ui/alert";

interface HackerNewsResponse {
    data: number[];
}
interface HackerNewsItem {
    id: number;
    title: string;
    url: string;
    score: number;
    by: string;
    time: number;
}

function App() {
    const [data, setData] = useState<HackerNewsResponse>({ data: [] });
    const [hackerItemData, setHackerItemData] = useState<HackerNewsItem[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState(false);

    useEffect(() => {
        const fetchData = async () => {
            setLoading(true);
            try {
                const response = await fetch(
                    "https://hacker-news.firebaseio.com/v0/newstories.json?print=pretty",
                );
                if (!response.ok && response.status !== 200) {
                    setError(true);
                    throw new Error("Invalid response");
                }
                const data = await response.json();
                console.log("Response:", data);
                setData({ data });
            } catch (error) {
                console.error("Error fetching data:", error);
                setError(true);
            } finally {
                setLoading(false);
            }
        };
        fetchData();
    }, []);

    useEffect(() => {
        if (!data || !data.data || data.data.length === 0) {
            return;
        }
        setLoading(true);
        const fetchItemData = async () => {
            const top20Ids = data.data.slice(0, 20);
            const result = await Promise.all(
                top20Ids.map(async (id) => {
                    try {
                        const res = await fetch(
                            `https://hacker-news.firebaseio.com/v0/item/${id}.json?print=pretty`,
                        );
                        if (!res.ok && res.status !== 200) {
                            throw new Error("Invalid response");
                        }
                        return res.json();
                    } catch (error) {
                        console.error(
                            `Error fetching item with id ${id}:`,
                            error,
                        );
                        setError(true);
                        return null;
                    }
                }),
            );
            // setHackerItemData([]);
            setHackerItemData(result.filter((item) => item !== null));
            setLoading(false);
        };

        fetchItemData();
    }, [data]);

    useEffect(() => {
        console.log("Hacker Item Data Updated:", hackerItemData);
    }, [hackerItemData]);

    const getRelativeTime = (timestamp: number) => {
        const diff = Date.now() - timestamp * 1000;
        const minutes = Math.floor(diff / 60000);

        if (minutes < 60) return `${minutes} min ago`;
        const hours = Math.floor(minutes / 60);
        if (hours < 24) return `${hours} hr ago`;
        const days = Math.floor(hours / 24);
        return `${days} days ago`;
    };

    return (
        <>
            <div className="w-full m-auto p-4">
                <h1 className="flex justify-center items-center h-full text-4xl font-bold">
                    React Async UI
                </h1>
            </div>

            {loading && (
                <div className="md:w-full xl:w-2/3 m-auto p-10">
                    <div>
                        <h1 className="text-xl font-bold pl-2">
                            List of Hacker News Post
                        </h1>
                    </div>
                    <Table>
                        <TableCaption>
                            A list of recent hacker news post.
                        </TableCaption>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="">Title</TableHead>
                                <TableHead>Score</TableHead>
                                <TableHead>By</TableHead>
                                <TableHead className="text-right">
                                    Time
                                </TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {Array.from({ length: 20 }, (_, i) => (
                                <TableRow key={i}>
                                    <TableCell>
                                        <div className="w-full m-auto">
                                            <Skeleton className="h-8 w-full" />
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="w-full m-auto">
                                            <Skeleton className="h-8 w-full" />
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="w-full m-auto">
                                            <Skeleton className="h-8 w-full" />
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="w-full m-auto">
                                            <Skeleton className="h-8 w-full" />
                                        </div>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                        <TableFooter></TableFooter>
                    </Table>
                </div>
            )}

            {!loading && error && (
                <div className="w-full flex justify-center items-center py-20">
                    <div className="max-w-md w-full">
                        <Alert variant="destructive">
                            <AlertCircle className="h-4 w-4" />
                            <AlertTitle>Something went wrong</AlertTitle>
                            <AlertDescription>
                                Failed to load Hacker News stories. Please try
                                again.
                            </AlertDescription>
                        </Alert>

                        <div className="flex justify-center mt-4">
                            <Button onClick={() => window.location.reload()}>
                                Retry
                            </Button>
                        </div>
                    </div>
                </div>
            )}

            {!loading && hackerItemData.length === 0 && !error && (
                <div className="w-full flex justify-center items-center py-20">
                    <div className="flex flex-col items-center text-center gap-4 max-w-sm">
                        <div className="bg-muted p-4 rounded-full">
                            <FileText className="w-6 h-6 text-muted-foreground" />
                        </div>

                        <h2 className="text-xl font-semibold">
                            No stories found
                        </h2>

                        <p className="text-sm text-muted-foreground">
                            Nothing to show right now. Try refreshing the page.
                        </p>

                        <Button onClick={() => window.location.reload()}>
                            Refresh
                        </Button>
                    </div>
                </div>
            )}

            {!loading && hackerItemData.length > 0 && (
                <div className="md:w-full xl:w-2/3 m-auto p-10">
                    <div>
                        <h1 className="text-xl font-bold pl-2">
                            List of Hacker News Post
                        </h1>
                    </div>
                    <Table className="table-fixed w-full">
                        <TableCaption className="text-muted-foreground">
                            Top Hacker News stories
                        </TableCaption>

                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[60%]">Story</TableHead>
                                <TableHead className="w-20">Score</TableHead>
                                <TableHead className="w-32">Author</TableHead>
                                <TableHead className="w-28 text-right">
                                    Time
                                </TableHead>
                            </TableRow>
                        </TableHeader>

                        <TableBody>
                            {hackerItemData.map((item) => (
                                <TableRow
                                    key={item.id}
                                    className="hover:bg-muted/50 transition-colors cursor-pointer"
                                    onClick={() =>
                                        window.open(
                                            item.url ||
                                                `https://news.ycombinator.com/item?id=${item.id}`,
                                            "_blank",
                                        )
                                    }
                                >
                                    <TableCell className="w-[60%]">
                                        <div className="flex flex-col gap-1">
                                            <span className="font-medium truncate">
                                                {item.title}
                                            </span>

                                            <span className="text-xs text-muted-foreground">
                                                {item.url
                                                    ? new URL(
                                                          item.url,
                                                      ).hostname.replace(
                                                          "www.",
                                                          "",
                                                      )
                                                    : "news.ycombinator.com"}
                                            </span>
                                        </div>
                                    </TableCell>

                                    <TableCell>
                                        <span className="font-medium">
                                            {item.score}
                                        </span>
                                    </TableCell>

                                    <TableCell>
                                        <span className="text-muted-foreground">
                                            {item.by}
                                        </span>
                                    </TableCell>

                                    <TableCell className="text-right text-muted-foreground">
                                        {getRelativeTime(item.time)}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </div>
            )}
        </>
    );
}

export default App;
