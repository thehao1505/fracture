# Implementation Approach

## Key Principles

Start with data, not UI: API → Hooks → Components
One responsibility per file: Don't mix concerns
Props down, events up: Data flows down, actions bubble up
Colocate related code: Keep feature code together
Build incrementally: Get one slice working end-to-end before adding complexity

## Development Order

Following bottom-up for back-end and top-down for front-end:

✅ Define types & interfaces
✅ Build API layer
✅ Create data fetching hooks
✅ Build UI components (pure/presentational)
✅ Create container/page component
✅ Add error handling & loading states
✅ Optimize performance