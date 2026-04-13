import type { PropsWithChildren, ReactNode } from "react";
import { render } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Route, Routes } from "react-router-dom";

function createTestQueryClient(): QueryClient {
    return new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
            },
        },
    });
}

interface RenderWithProvidersOptions {
    route?: string;
    path?: string;
}

export function renderWithProviders(
    ui: ReactNode,
    { route = "/", path }: RenderWithProvidersOptions = {},
) {
    const queryClient = createTestQueryClient();

    function Wrapper({ children }: PropsWithChildren) {
        return (
            <MemoryRouter initialEntries={[route]}>
                <QueryClientProvider client={queryClient}>
                    {path ? (
                        <Routes>
                            <Route path={path} element={<>{children}</>} />
                        </Routes>
                    ) : (
                        children
                    )}
                </QueryClientProvider>
            </MemoryRouter>
        );
    }

    return {
        queryClient,
        ...render(<>{ui}</>, { wrapper: Wrapper }),
    };
}
