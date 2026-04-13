# React Async UI

A React + TypeScript project that explores async UI patterns with Hacker News data.

The app lets users browse stories, search for posts, compare manual `fetch`-based data loading with React Query, and open a dedicated details view for each story. It is built to demonstrate practical handling of loading, error, empty, and success states in a responsive interface.

## Overview

This project was built as an async data-fetching exercise using public Hacker News APIs. It focuses on the parts of frontend work that are easy to get wrong in real applications:

- rendering skeletons while data is loading
- showing clear retry flows when requests fail
- handling empty results gracefully
- preventing stale search results with request cancellation
- comparing raw `fetch` logic with React Query caching
- testing the important user-facing states with Vitest

## Features

- Browse top Hacker News stories on the homepage
- Search stories using the Algolia Hacker News API
- Toggle between `Fetch API` and `React Query` loading strategies
- View loading skeletons during async requests
- See friendly error and empty states with clear recovery actions
- Open a story details page with metadata and external link
- Navigate between list and details pages with React Router
- Responsive layout for mobile, tablet, and desktop screens
- Tailwind CSS styling
- Unit tests with coverage thresholds enabled

## Tech Stack

- React 19
- TypeScript
- Vite
- Tailwind CSS
- React Router
- TanStack React Query
- Vitest
- Testing Library

## Demo Behavior

The homepage supports two data-loading modes:

- `Fetch API`: manual request handling with local state and cached in-memory results
- `React Query`: request management and caching handled through TanStack Query

Search uses the Algolia Hacker News API and passes an `AbortSignal` so previous in-flight requests can be cancelled when new input arrives.

## Project Structure

```text
src/
  api/
    hackerNewsApi.ts         # API calls and response mapping
    hackerNewsApi.test.ts    # API unit tests
  pages/
    HomePage.tsx             # Story list, search, async states
    HomePage.test.tsx        # Homepage interaction tests
    DetailsPage.tsx          # Story details page
    DetailsPage.test.tsx     # Details page tests
  test/
    setup.ts                 # Vitest setup
    test-utils.tsx           # Shared render helpers
  type/
    types.ts                 # Shared TypeScript types
  util/
    utils.ts                 # Helper functions
    utils.test.ts            # Utility tests
  App.tsx                    # Route configuration
  main.tsx                   # App bootstrap
```

## Getting Started

### Prerequisites

- Node.js 20+
- pnpm 9+

### Installation

```bash
pnpm install
```

### Start the development server

```bash
pnpm dev
```

The app will be available at `http://localhost:5173` by default.

## Available Scripts

```bash
pnpm dev           # Start the Vite dev server
pnpm build         # Type-check and build for production
pnpm preview       # Preview the production build locally
pnpm lint          # Run ESLint
pnpm test          # Run tests once
pnpm test:watch    # Run tests in watch mode
pnpm test:coverage # Run tests with coverage output
```

## Testing

The project uses Vitest and Testing Library for unit and component-level testing.

Current coverage thresholds in `vitest.config.ts`:

- Statements: 80%
- Branches: 80%
- Functions: 80%
- Lines: 80%

The tests cover key async UI behavior such as:

- loading states
- error states
- empty states
- successful rendering
- search interactions
- navigation behavior
- API response mapping

## Routing

The application uses two main routes:

- `/` for the homepage story list and search UI
- `/post/:id` for the story details page

## Data Sources

- Hacker News Firebase API: `https://hacker-news.firebaseio.com/v0`
- Algolia Hacker News Search API: `https://hn.algolia.com/api/v1`

## Async UI Notes

This codebase intentionally demonstrates multiple async UI patterns in one place.

- Manual fetching is handled with `useEffect`, local state, and an in-memory cache
- React Query handles caching and async state for the alternate flow
- Search requests pass through an abort signal to reduce stale-result issues
- Error and empty states are rendered directly in the page UI so the user always gets clear feedback

## Styling

Tailwind CSS is used for all styling. The UI is designed to stay readable and usable across different viewport sizes with simple responsive layout choices.

## Build Notes

- `src/main.tsx` is excluded from coverage reporting
- The project does not require environment variables to run locally
- No backend service is needed because all data comes from public APIs

## Future Improvements

- Add integration or end-to-end tests for full route flows
- Improve accessibility coverage for keyboard and screen-reader interactions
- Add pagination or infinite scrolling for larger result sets
- Add richer story metadata and comments support on the details page

## License

This project is for learning and assessment purposes.
