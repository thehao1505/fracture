# Next.js API Route Best Practices

## File Structure & Organization

- Use the App Router's route handlers (`app/api/[route]/route.ts`) for new projects instead of Pages Router API routes
- Organize API routes logically by feature or domain (e.g., `app/api/users/route.ts`, `app/api/products/route.ts`)
- Keep route handlers focused on a single responsibility and extract complex business logic into separate service files
- Use dynamic route segments for RESTful patterns (e.g., `app/api/users/[id]/route.ts` for individual user operations)

## HTTP Methods & RESTful Design

- Export named functions for each HTTP method: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`
- Return appropriate HTTP status codes: 200 for success, 201 for created resources, 400 for bad requests, 401 for unauthorized, 404 for not found, 500 for server errors
- Use `NextResponse.json()` to return JSON responses with proper status codes
- Implement proper error handling with try-catch blocks and return meaningful error messages

## Request Handling

- Access request data using the `Request` object parameter: `await request.json()` for body, `request.nextUrl.searchParams` for query parameters
- Validate all incoming request data before processing using libraries like Zod or Yup
- Sanitize user inputs to prevent security vulnerabilities like XSS and SQL injection
- Use TypeScript to define request and response types for better type safety

## Response Formatting

- Maintain consistent response structure across all endpoints (e.g., `{ data, error, message }`)
- Always include proper Content-Type headers in responses
- Use `NextResponse.json()` for JSON responses and set appropriate headers
- Return pagination metadata for list endpoints (e.g., `total`, `page`, `pageSize`, `hasMore`)