---
agent: agent
model: Claude Opus 4.8
tools: [execute, read, agent, edit, todo]
---

Next.js TypeScript Project Coding Knowledge Base Generation Workflow

<roleContext>
YOU ARE an expert Next.js and TypeScript architect specialized in modern web application analysis and technical documentation.
THIS WORKFLOW: Analyzes this Next.js 14 TypeScript codebase with App Router architecture and generates a comprehensive knowledge.coding.md document serving as a technical onboarding guide by delivering systematic Next.js project structure analysis with React patterns, TypeScript conventions, and Next.js-specific design choices with absolute precision.
</roleContext>

<objectives>
<primary>Generate a comprehensive knowledge.coding.md document that serves as a technical onboarding guide for Next.js TypeScript development and React architecture</primary>
<secondary>
    <goal>Document Next.js App Router patterns, API routes structure, and React component architecture with absolute precision</goal>
    <goal>Extract and document Next.js development workflow commands and TypeScript build processes for team onboarding</goal>
    <goal>Map React component hierarchies, state management patterns, and Next.js-specific functionalities</goal>
    <goal>Create objective technical documentation focused on Next.js project structure and modern React development practices</goal>
</secondary>
</objectives>

<importantReminders>USE your Todo Management tool to track task progress throughout this entire workflow execution.</importantReminders>

<executionFlow>
WORKFLOW METHODOLOGY:
1. VALIDATE pre-workflow tasks
2. EXECUTE workflow phases sequentially
3. INTEGRATE post-workflow tasks
</executionFlow>

<preWorkflowTasks>
BEFORE STARTING: EXECUTE these validation and setup tasks in sequence. STOP and Report if any task fails:
    <task title="Create Empty Knowledge Files">
        mkdir -p .github/docs && \
        touch .github/docs/knowledge.coding.md
    </task>
</preWorkflowTasks>

<phases>
EXECUTE the following phases SEQUENTIALLY. COMPLETE each phase entirely before proceeding to the next:
    <phase number="1" name="Next.js Architecture Assessment">
        <task id="1.1" title="Next.js App Router Structure Analysis">
            ANALYZE the src/app directory structure using Next.js 14 App Router conventions. DETERMINE the routing patterns, layout hierarchy, and page organization. DOCUMENT the purpose of ALL route groups, parallel routes, and intercepting routes if present.
        </task>
        <task id="1.2" title="TypeScript Configuration and Code Organization">
            INVESTIGATE TypeScript configuration in tsconfig.json, path aliases, and type definitions. IDENTIFY and DOCUMENT the established patterns for component types, service interfaces, model definitions, and shared TypeScript utilities in src/types, src/models, and src/utils.
        </task>
        <task id="1.3" title="React Component Architecture Analysis">
            ANALYZE the src/components directory structure and component organization patterns. DOCUMENT the component hierarchy, reusable component patterns, and the relationship between feature-specific components in src/features and shared components.
        </task>
        <task id="1.4" title="Cross-Cutting Concerns in Next.js Context">
            FIND ALL implementations for Next.js-specific functionalities. DOCUMENT the project's approach to middleware, authentication (src/lib/auth.ts), API route organization (src/app/api), state management patterns (MobX usage), and client-server data flow.
        </task>
        <task id="1.5" title="Update knowledge.coding.md">
           UPDATE the knowledge.coding.md file with the documented Next.js architectural assessment information.
        </task>
    </phase>
    <phase number="2" name="Feature and Service Layer Analysis">
        <task id="2.1" title="Feature Module Architecture">
            ANALYZE each feature module in src/features directory. DOCUMENT the internal structure pattern of each feature including view models, components, and their relationship to the overall application architecture.
        </task>
        <task id="2.2" title="Service Layer and API Integration">
            ANALYZE the src/services directory and API integration patterns. DOCUMENT the service layer architecture, authentication handling in fetchWithAuth.ts, and the relationship between frontend services and API routes.
        </task>
        <task id="2.3" title="State Management and Data Flow">
            INVESTIGATE MobX integration and state management patterns. DOCUMENT how view models, services, and React components interact for data flow and state management across the application.
        </task>
        <task id="2.4" title="Update knowledge.coding.md">
           UPDATE the knowledge.coding.md file with the Feature and Service Layer Analysis.
        </task>
    </phase>
    <phase number="3" name="Infrastructure and Development Workflow">
        <task id="3.1" title="Next.js API Routes and Data Schema">
            IDENTIFY all API routes in src/app/api and their purposes. ANALYZE type definitions in src/types to understand data models. DOCUMENT the API structure and data flow patterns between frontend and backend.
        </task>
        <task id="3.2" title="Next.js Development Workflow Extraction">
            ANALYZE package.json scripts and Next.js configuration. LIST and DESCRIBE ALL essential commands for: dependency installation with pnpm, development server startup, production builds, linting with next lint, and TypeScript type checking.
        </task>
        <task id="3.3" title="Styling and UI Framework Integration">
            DOCUMENT the styling approach including Tailwind CSS configuration, Ant Design integration with Next.js 14, custom component styling patterns, and responsive design implementation.
        </task>
        <task id="3.4" title="Build and Deployment Configuration">
            ANALYZE next.config.mjs, Dockerfile, and docker-compose.yaml. DOCUMENT the build configuration, deployment settings, standalone output configuration, and containerization setup.
        </task>
        <task id="3.5" title="Update knowledge.coding.md">
           UPDATE the knowledge.coding.md file with the Infrastructure and Development Workflow information.
        </task>
    </phase>
</phases>

<postWorkflowTasks>
AFTER COMPLETING all phases: EXECUTE these tasks:
    <task title="Execute Reflection Workflow">
        EXECUTE and READ `cat .github/prompts/reflections/a.onboarding.coding.reflection.md` workflow
    </task>
</postWorkflowTasks>

<constraints>
ABSOLUTE RESTRICTIONS - NEVER violate these rules:
- NEVER include business-specific logic details or domain-specific feature implementations
- NEVER count or enumerate specific API endpoints or business features
- NEVER include implementation details unrelated to Next.js architecture and React patterns
- ALWAYS focus ONLY on Next.js structural patterns, TypeScript conventions, and React architectural decisions
- MUST maintain objective, technical tone throughout documentation focused on Next.js and React development practices
- FOLLOW the project's established rule: Do not write any comments in code examples (no TODO, FIXME, function, class, interface comments)
</constraints>

<executionInstructions>
<command>**EXECUTE NOW**: Begin autonomous execution of ALL tasks following the methodology above.</command>
<autonomyLevel>Full autonomous execution with ONLY HIGH-LEVEL progress reporting.</autonomyLevel>
</executionInstructions>
