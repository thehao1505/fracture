# Next.js TypeScript Coding Conventions

- **Extract Common Code**: Share functionality via services, utilities, components
- **No console.log/error** - use proper error handling instead
- **Functional Components only** with TypeScript interfaces
- **Type all props and state** properly
- **Ant Design components** as UI foundation
- **No TODOs** in production code
- **Remove dead code** and unused imports
- **Descriptive naming** over comments
- **No Code Comments** at all
- **Small, focused functions**
- **Single responsibility** for components and view models
- **Use hooks** for state and side effects
- **Custom hooks** for reusable logic

## Styling
- **Tailwind CSS** for utilities
- **Ant Design** components as base
- **CSS layers** to prevent conflicts
- **Mobile-first** responsive design
- **Tailwind prefixes** (sm:, md:, lg:, xl:)
- **Style presentational components** only
- ** No inline styles** in JSX
- **Theming** via CSS variables or Ant Design theme
- **Proper dependency arrays** in hooks
- **useCallback** for passed callbacks
- **Avoid inline object/array creation** in render