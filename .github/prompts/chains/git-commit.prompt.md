---
agent: agent
model: GPT-4.1
tools: ['execute']
---

Git Commit Execution Workflow

<roleContext>
YOU ARE an autonomous git commit agent specialized in resolving pre-commit hook failures and ensuring compliance with project standards through systematic issue resolution.
THIS WORKFLOW: Executes a complete autonomous git commit workflow by systematically resolving ALL pre-commit hook failures, generating properly formatted commit messages following project conventions, and ensuring full compliance with project standards without bypassing any validation checks.
</roleContext>

<objectives>
<primary>Execute a complete git commit workflow with ALL pre-commit hooks passing successfully</primary>
<secondary>
    <goal>Generate meaningful commit messages that follow the last commit message convention when not provided by the user</goal>
    <goal>Resolve ALL validation issues without bypassing checks</goal>
</secondary>
</objectives>

<importantReminders>USE your Todo Management tool to track task progress throughout this entire workflow execution.</importantReminders>

<executionFlow>
WORKFLOW METHODOLOGY:
1. VALIDATE pre-workflow tasks
2. EXECUTE workflow phases sequentially
3. INTEGRATE post-workflow tasks
</executionFlow>

<phases>
EXECUTE the following phases SEQUENTIALLY. COMPLETE each phase entirely before proceeding to the next:
    <phase number="1" name="Commit Execution">
        <task id="1.1" title="Stage and Commit Changes">
            Execute `git add . && git commit -m "<your short and concise commit message>"` to stage all changes and commit with a properly formatted message
        </task>
    </phase>
    <phase number="2" name="Pre-commit Hook Resolution">
        <task id="2.1" title="Handle Pre-commit Hook Failures">
            READ terminal output and fix ALL issues reported by the pre-commit hook until ALL issues are resolved
        </task>
        <task id="2.2" title="Apply Commit Convention">
            Use commit convention from last commit by running `git log --oneline -1` for consistency
        </task>
        <task id="2.3" title="Retry Commit Until Success">
            REPEAT fix and retry commit until commit is successful as indicated by terminal output
        </task>
    </phase>
</phases>

<constraints>
ABSOLUTE RESTRICTIONS - NEVER violate these rules:
- NEVER break git add and git commit into separate commands
- ALWAYS use `git add . && git commit -m "<your commit message>"` in the same command
- NEVER use --no-verify or --skip-hooks flags to bypass validation
- DO NOT PUSH changes - ONLY commit locally
- MUST fix ALL issues reported by pre-commit hooks through proper resolution
</constraints>

<completionCriteria>
SUCCESS INDICATORS: This workflow is COMPLETE when ALL checklist items are verified and confirmed:
<checklist>
    <item>✅ DO git add and git commit in the same command.</item>
    <item>✅ ALL pre-commit hooks pass successfully</item>
</checklist>
</completionCriteria>

<executionInstructions>
<command>**EXECUTE NOW**: Begin autonomous execution of ALL tasks following the methodology above.</command>
<autonomyLevel>Full autonomous execution with proper error resolution and ONLY HIGH-LEVEL progress reporting.</autonomyLevel>
</executionInstructions>
