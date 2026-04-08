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

    useEffect(() => {
        const fetchData = async () => {
            setLoading(true);
            try {
                const response = await fetch(
                    "https://hacker-news.firebaseio.com/v0/newstories.json?print=pretty",
                );
                const data = await response.json();
                console.log("Response:", data);
                setData({ data });
            } catch (error) {
                console.error("Error fetching data:", error);
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
                        return res.json();
                    } catch (error) {
                        console.error(
                            `Error fetching item with id ${id}:`,
                            error,
                        );
                        return null;
                    }
                }),
            );
            // setHackerItemData([]);
            setHackerItemData(result.filter((item) => item !== null));
            // for (const id of top20Ids) {
            //     try {
            //         const itemResponse = await fetch(
            //             `https://hacker-news.firebaseio.com/v0/item/${id}.json?print=pretty`,
            //         );
            //         const hackerNewsItem = await itemResponse.json();

            //         setHackerItemData((prevData) => [
            //             ...prevData,
            //             hackerNewsItem,
            //         ]);
            //     } catch (error) {
            //         console.error(`Error fetching item with id ${id}:`, error);
            //     }
            // }
            // console.log("Hacker Item Data:", hackerItemData);
            setLoading(false);
        };

        fetchItemData();
    }, [data]);

    useEffect(() => {
        console.log("Hacker Item Data Updated:", hackerItemData);
    }, [hackerItemData]);

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
            {!loading && hackerItemData.length === 0 && (
                <div className="w-full m-auto p-4">
                    <h2 className="flex justify-center items-center h-full text-2xl font-semibold">
                        No Hacker News posts available.
                    </h2>
                </div>
            )}

            {!loading && hackerItemData.length > 0 && (
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
                            {hackerItemData.map((item) => (
                                <TableRow key={item.id}>
                                    <TableCell className="font-medium max-w-28 md:max-w-52 lg:max-w-80 xl:max-w-lg truncate">
                                        {item.title}
                                    </TableCell>
                                    <TableCell>{item.score}</TableCell>
                                    <TableCell>{item.by}</TableCell>
                                    <TableCell className="text-right">
                                        {new Date(item.time).toLocaleString()}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                        <TableFooter></TableFooter>
                    </Table>
                </div>
            )}
        </>
    );
}

export default App;
