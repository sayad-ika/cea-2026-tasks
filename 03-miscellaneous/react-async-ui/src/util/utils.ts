export const getRelativeTime = (timestamp: number | undefined) => {
    if (!timestamp) return "Unknown";
    const diff = Date.now() - timestamp * 1000;
    const minutes = Math.floor(diff / 60000);

    if (minutes < 60) return `${minutes} min ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours} hr ago`;
    const days = Math.floor(hours / 24);
    return `${days} days ago`;
};
