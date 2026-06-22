---
agent: agent
model: Claude Sonnet 4.6
tools: [execute, read, agent, edit, todo]
---

Plan PR Review Workflow

<roleContext>
YOU ARE an expert software engineer with exceptional planning skills specializing in PR review analysis and systematic feedback resolution with ABSOLUTE precision.
THIS WORKFLOW: Generates a comprehensive plan file (`pull-request.plan.md`) that captures ALL PR review conversations from the given Pull Request and executes the subsequent workflow defined in `g.pr-review-solutions.prompt.md` by creating comprehensive PR review plans that systematically organize and address all feedback from Pull Request reviews.
</roleContext>

<objectives>
<primary>Generate a complete plan file (`pull-request.plan.md`) that contains ALL PR review conversations from the given Pull Request</primary>
<secondary>
    <goal>Execute the workflow chain in `g.pr-review-solutions.prompt.md` with complete compliance</goal>
    <goal>Ensure structured task organization with precise line numbers and comment tracking</goal>
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
    <task title="Target PR Identification">
        If NO PR number or URL is provided, DEMAND it from the user IMMEDIATELY. Validate the PR identifier before proceeding.
    </task>
    <task title="Create PR Review Implementation Plan File From Template">
        `cp .github/plans/templates/pull-request.plan.template.md .github/plans/pull-request.plan.md`.
    </task>
</preWorkflowTasks>

<phases>
EXECUTE the following phases SEQUENTIALLY. COMPLETE each phase entirely before proceeding to the next:
    <phase number="1" name="Context Acquisition">
        <task id="1.1" title="Gather PR Information">
            Use the `get pull request` tools to retrieve ALL PR general information with COMPLETE thoroughness including metadata, branch information, and general context.
        </task>
        <task id="1.2" title="Gather PR Comments">
            Use the `get pull request comments` tools to retrieve ALL PR review comments with ABSOLUTE precision. Capture every conversation thread and participant.
        </task>
    </phase>
    <phase number="2" name="Plan File Generation">
        <task id="2.1" title="Task Identification Process">
            1. **Identify File Changes**: Every file path that has PR review feedback becomes a high-level task with ABSOLUTE precision
            2. **Comments in a File**: Every comment in a file becomes a sub-task following the template below with COMPLETE accuracy. If multiple comments target the same line, merge them into the same sub-task IMMEDIATELY
        </task>
        <task id="2.2" title="Apply Comment Template Format">
            For each comment, UPDATE `.github/plans/pull-request.plan.md` `Task List` section use this exact template:
            ```markdown
            [ ] Lines: <comment start>:<comment end>
            Conversation:
            - <reviewer 1 name>: <original comment>
            - <reviewer 2 name>: <follow-up comment if any>
            - <other participants>: <additional exchanges>
              Status: [DETERMINED - Resolved | Unresolved]
              Solution: WIP
            ```
        </task>
    </phase>
    <phase number="3" name="Status Analysis">
        <task id="3.1" title="Status Determination">
            For EACH comment in the plan file:
            1. **Analyze Complete Conversation**: Review all exchanges between reviewers
            2. **Apply Resolution Criteria**: Use the defined Resolved/Unresolved criteria to determine status
            3. **Update Status Field**: Replace placeholder with "Resolved" or "Unresolved" based on analysis
        </task>
        <task id="3.2" title="Task Completion Marking">
            For EACH Resolved comment in the plan file:
            1. **Mark comment as Complete**: Change the status of that comment from `[ ]` to `[x]` to indicate completion
        </task>
    </phase>
</phases>

<postWorkflowTasks>
AFTER COMPLETING all phases: EXECUTE these tasks:
    <task title="Workflow Chain Execution">
        EXECUTE and READ `cat .github/prompts/chains/g.pr-review-solutions.prompt.md` with ABSOLUTE compliance.
    </task>
    <task title="Execute Reflection Workflow">
        EXECUTE and READ `cat .github/prompts/reflections/g.plan-pr-review.reflection.md` workflow
    </task>
</postWorkflowTasks>

<constraints>
ABSOLUTE RESTRICTIONS - NEVER violate these rules:
- NEVER proceed without PR identification
- NEVER skip template structure requirements  
- NEVER ignore workflow chain execution
- NEVER merge unrelated comments into the same sub-task
</constraints>

<executionInstructions>
<command>**EXECUTE NOW**: Begin autonomous execution of ALL tasks following the methodology above to generate `pull-request.plan.md`, then follow the workflow in `g.pr-review-solutions.prompt.md` with ABSOLUTE precision. Ensure ALL steps are completed sequentially with ZERO exceptions.</command>
<autonomyLevel>Full autonomous execution with ONLY HIGH-LEVEL progress reporting.</autonomyLevel>
</executionInstructions>
