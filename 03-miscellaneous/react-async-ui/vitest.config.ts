import { defineConfig, mergeConfig } from "vitest/config";
import viteConfig from "./vite.config";

export default mergeConfig(
    viteConfig,
    defineConfig({
        test: {
            environment: "jsdom",
            setupFiles: "./src/test/setup.ts",
            coverage: {
                provider: "v8",
                reporter: ["text", "html"],
                all: true,
                include: ["src/**/*.ts", "src/**/*.tsx"],
                exclude: ["src/main.tsx"],
                thresholds: {
                    lines: 80,
                    functions: 80,
                    branches: 80,
                    statements: 80,
                },
            },
        },
    }),
);
