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

    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await fetch(
                    "https://hacker-news.firebaseio.com/v0/newstories.json?print=pretty",
                );
                const data = await response.json();
                console.log("Response:", data);
                setData({ data });
            } catch (error) {
                console.error("Error fetching data:", error);
            }
        };
        fetchData();
    }, []);

    useEffect(() => {
        if (!data || !data.data || data.data.length === 0) {
            return;
        }

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
            console.log("Hacker Item Data:", hackerItemData);
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
            <div className="w-1/2 m-auto p-10">
                <Table>
                    <TableCaption>
                        A list of recent hacker news post.
                    </TableCaption>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-25">Title</TableHead>
                            <TableHead>Score</TableHead>
                            <TableHead>By</TableHead>
                            <TableHead className="text-right">Time</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {hackerItemData.map((item) => (
                            <TableRow key={item.id}>
                                <TableCell className="font-medium">
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
        </>
    );
}

export default App;
