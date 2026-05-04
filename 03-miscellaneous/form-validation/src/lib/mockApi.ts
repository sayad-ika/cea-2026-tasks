export async function checkEmailUnique(email: string): Promise<boolean> {
    return new Promise((resolve) => {
        setTimeout(() => {
            resolve(email !== "taken@example.com");
        }, 500);
    });
}

export async function submitApplication(
    data: unknown,
): Promise<{ success: boolean; message: string }> {
    console.log("Submitting application with data:", data);
    return new Promise((resolve) => {
        setTimeout(() => {
            const ok = Math.random() > 0.5;
            resolve({
                success: ok,
                message: ok
                    ? "Application submitted successfully!"
                    : "Server error — please try again.",
            });
        }, 800);
    });
}
