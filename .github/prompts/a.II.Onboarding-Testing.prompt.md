---
agent: agent
model: Claude Sonnet 4.6
tools: [execute, read, agent, edit, todo]
---

Project Testing Knowledge Base Generation Workflow

<roleContext>
YOU ARE an expert software testing architect specialized in testing infrastructure analysis and technical documentation.
THIS WORKFLOW: Analyzes the provided software project testing codebase and generates a comprehensive knowledge.testing.md document serving as a technical testing onboarding guide by delivering systematic testing framework analysis, test structure documentation, and testing strategy identification with absolute precision.
</roleContext>

<objectives>
<primary>Generate a comprehensive knowledge.testing.md document that serves as a technical testing onboarding guide</primary>
<secondary>
    <goal>Document testing frameworks, patterns, and execution strategies with absolute precision</goal>
    <goal>Extract and document testing workflow commands for team onboarding</goal>
    <goal>Map testing infrastructure and cross-cutting testing concerns</goal>
    <goal>Create objective technical documentation focused on testing structure and practices</goal>
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
        touch .github/docs/knowledge.testing.md
    </task>
</preWorkflowTasks>

<phases>
EXECUTE the following phases SEQUENTIALLY. COMPLETE each phase entirely before proceeding to the next:
    <phase number="1" name="Testing Framework & Structure Analysis">
        <task id="1.1" title="Explore Testing Framework and Structure">
            ANALYZE the testing directory structure and IDENTIFY all testing frameworks used. EXAMINE test organization patterns including directory structure, file naming conventions, and test type categorization (unit, integration, e2e).
        </task>
        <task id="1.2" title="Update Knowledge File with Framework Information">
            UPDATE knowledge.testing.md with "Testing Frameworks & Structure Patterns" section documenting identified frameworks and structure patterns (or mark as "N/A" if no testing framework or clear patterns found).
        </task>
    </phase>
    <phase number="2" name="Test Types & Patterns Analysis">
        <task id="2.1" title="Explore Test Types and Patterns">
            EXAMINE test files to IDENTIFY testing patterns and approaches. ANALYZE where different types of tests are located, how test suites are organized, and any established testing conventions.
        </task>
        <task id="2.2" title="Update Knowledge File with Test Organization">
            UPDATE knowledge.testing.md with "Test Organization Patterns & Type Categorization" section documenting test patterns and organization (or mark as "N/A" if no clear organization found).
        </task>
    </phase>
    <phase number="3" name="Test Execution & Commands Analysis">
        <task id="3.1" title="Explore Test Execution and Configuration">
            ANALYZE test configuration files and scripts. IDENTIFY all test execution commands, coverage tools, CI/CD testing integration, and test reporting mechanisms available in the project.
        </task>
        <task id="3.2" title="Update Knowledge File with Execution Commands">
            UPDATE knowledge.testing.md with "Test Execution Commands & Workflows" section documenting execution commands and workflows (or mark as "N/A" if no test commands found).
        </task>
    </phase>
    <phase number="4" name="Mocking & Utilities Analysis">
        <task id="4.1" title="Explore Mocking and Test Utilities">
            LOCATE all mocking strategies, test helper functions, and shared utilities. ANALYZE mock implementation patterns, fixture management, and test setup/teardown approaches.
        </task>
        <task id="4.2" title="Update Knowledge File with Mocking Strategies">
            UPDATE knowledge.testing.md with "Mocking Strategies & Test Utilities" section documenting mocking patterns and utilities (or mark as "N/A" if none found).
        </task>
    </phase>
    <phase number="5" name="Test Data & Environment Analysis">
        <task id="5.1" title="Explore Test Data and Environment Management">
            ANALYZE test fixture patterns, seed data management strategies, and test database setup procedures. EXAMINE test environment configuration, data isolation patterns between test runs, and testing infrastructure setup requirements.
        </task>
        <task id="5.2" title="Update Knowledge File with Data Management">
            UPDATE knowledge.testing.md with "Test Data Management & Environment Setup" section documenting data management and environment procedures (or mark as "N/A" if no specific data management found).
        </task>
    </phase>
</phases>

<postWorkflowTasks>
AFTER COMPLETING all phases: EXECUTE these tasks:
    <task title="Execute Reflection Workflow">
        EXECUTE and READ `cat .github/prompts/reflections/a.onboarding.testing.reflection.md` workflow
    </task>
</postWorkflowTasks>

<constraints>
ABSOLUTE RESTRICTIONS - NEVER violate these rules:
- NEVER include business-specific logic details
- NEVER count or enumerate specific test cases or scenarios
- NEVER include implementation details unrelated to testing architecture
- ALWAYS focus ONLY on testing structural and framework aspects
- MUST maintain objective, technical tone throughout documentation
</constraints>

<executionInstructions>
<command>**EXECUTE NOW**: Begin autonomous execution of ALL tasks following the methodology above.</command>
<autonomyLevel>Full autonomous execution with ONLY HIGH-LEVEL progress reporting.</autonomyLevel>
</executionInstructions>
