import { useState } from "react";
import "./App.css";

const API_ENDPOINT = import.meta.env.VITE_API_ENDPOINT;

type Status = "idle" | "sending" | "success" | "error";

const CHARACTERS =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

function generateRandomId(): string {
    let result = "";
    for (let i = 0; i < 8; i++) {
        result += CHARACTERS.charAt(
            Math.floor(Math.random() * CHARACTERS.length),
        );
    }
    return result;
}

function getDefaultMessage(): string {
    const id = `msg-${generateRandomId()}`;
    return JSON.stringify(
        {
            id,
            type: "order",
            payload: { amount: 100, currency: "USD" },
        },
        null,
        2,
    );
}

function App() {
    const [message, setMessage] = useState<string>(getDefaultMessage);
    const [status, setStatus] = useState<Status>("idle");
    const [errorMsg, setErrorMsg] = useState<string>("");

    const handleSend = async () => {
        try {
            JSON.parse(message);
        } catch (err) {
            setStatus("error");
            setErrorMsg(
                err instanceof SyntaxError ? err.message : "Invalid JSON",
            );
            return;
        }

        setStatus("sending");
        setErrorMsg("");

        try {
            const response = await fetch(API_ENDPOINT, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: message,
            });

            if (response.ok) {
                setStatus("success");
                setMessage(getDefaultMessage());
            } else {
                setStatus("error");
                setErrorMsg(`${response.status} ${response.statusText}`);
            }
        } catch (err) {
            setStatus("error");
            setErrorMsg(err instanceof Error ? err.message : "Network error");
        }
    };

    return (
        <div className="page">
            <div className="card">
                <h1 className="card-title">SQS Message Publisher</h1>
                <p className="card-subtitle">
                    Send a JSON message to the queue via the API endpoint.
                </p>

                <div className="textarea-wrapper">
                    <textarea
                        className="textarea"
                        value={message}
                        onChange={(e) => {
                            setMessage(e.target.value);
                            if (status === "success" || status === "error") {
                                setStatus("idle");
                                setErrorMsg("");
                            }
                        }}
                        spellCheck={false}
                    />
                </div>

                <button
                    className="send-button"
                    type="button"
                    disabled={status === "sending"}
                    onClick={handleSend}
                >
                    {status === "sending" ? "Sending\u2026" : "Send Message"}
                </button>

                <div className="status-area">
                    {status === "success" && (
                        <span className="status-pill success">
                            <svg
                                className="status-icon"
                                viewBox="0 0 16 16"
                                fill="none"
                                xmlns="http://www.w3.org/2000/svg"
                            >
                                <path
                                    d="M4 8.5L7 11.5L12 4.5"
                                    stroke="currentColor"
                                    strokeWidth="2"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                />
                            </svg>
                            Message sent successfully
                        </span>
                    )}
                    {status === "error" && (
                        <span className="status-pill error">
                            <svg
                                className="status-icon"
                                viewBox="0 0 16 16"
                                fill="none"
                                xmlns="http://www.w3.org/2000/svg"
                            >
                                <path
                                    d="M4 4L12 12M12 4L4 12"
                                    stroke="currentColor"
                                    strokeWidth="2"
                                    strokeLinecap="round"
                                />
                            </svg>
                            <span className="error-text">{errorMsg}</span>
                        </span>
                    )}
                </div>
            </div>
        </div>
    );
}

export default App;
