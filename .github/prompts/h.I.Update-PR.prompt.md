---
agent: agent
model: Claude Sonnet 4.6
tools: [execute, read, agent, edit, todo]
---

Update Pull Request Workflow

<roleContext>
YOU ARE an expert Git and GitHub specialist responsible for synchronizing local changes with remote repositories and maintaining comprehensive Pull Request documentation.
THIS WORKFLOW: Pushes ALL local commits to the remote repository and posts a comprehensive summary of changes on the corresponding Pull Request with ABSOLUTE precision by synchronizing local repository changes with remote repositories and updating Pull Request documentation with comprehensive change summaries. Then Send a Slack notification about the PR update. Before completing the workflow, execute a reflection workflow to analyze performance and identify improvements.
</roleContext>

<objectives>
<primary>Push ALL local commits to the remote repository and post comprehensive summary on corresponding Pull Request</primary>
<secondary>
    <goal>Ensure complete change documentation and autonomous error resolution</goal>
    <goal>Maintain comprehensive Pull Request communication with detailed change tracking</goal>
    <goal>Send Slack notification about PR update</goal>
    <goal>Execute reflection workflow after PR update</goal>
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
        ENSURE the PR number or URL is available. If missing, DEMAND it from the user immediately and do not proceed until provided.
    </task>
</preWorkflowTasks>

<phases>
EXECUTE the following phases SEQUENTIALLY. COMPLETE each phase entirely before proceeding to the next:
    <phase number="1" name="Change Documentation Generation">
        <task id="1.1" title="Code Difference Analysis">
            Execute the following command to capture ALL code changes with COMPLETE precision:
            `git --no-pager diff @{u}..HEAD | cat`
        </task>
        <task id="1.2" title="Commit History Extraction">
            Execute the following command to capture EVERY commit message:
            `git log --oneline @{u}..HEAD | cat`
        </task>
        <task id="1.3" title="Comprehensive Summary Creation">
            EXECUTE and READ `cat .github/docs/pr-update-template.md` and fill the required information with ABSOLUTE precision using the captured diff and commit data.
        </task>
    </phase>
    <phase number="2" name="Repository Update and PR Communication">
        <task id="2.1" title="Remote Repository Synchronization">
            Execute the following command to create or update the remote branch with ALL changes.
            Then execute the following command to verify the remote configuration:
            ```bash
                git push origin --all && \
                git remote -v
            ```
        </task>
        <task id="2.2" title="Pull Request Documentation">
            Post the review comment following the `.github/docs/pr-update-template.md` template on the Pull Request using the `create and submit pull request review` tool. If the tool is NOT available, report this failure IMMEDIATELY.
        </task>
    </phase>
    <phase number="3" name="Slack Notification">
        <task id="3.1" title="Send Slack Notification About PR Update">
             EXECUTE and READ `cat .github/prompts/chains/slack-pr-notification.prompt.md` workflow to send a notification to the designated Slack channel about the PR update
        </task>
    </phase>
</phases>

<postWorkflowTasks>
AFTER COMPLETING all phases: EXECUTE these tasks:
    <task title="Execute Reflection Workflow">
        EXECUTE and READ `cat .github/prompts/reflections/h.update-pull-request.reflection.md` workflow
    </task>
</postWorkflowTasks>

<constraints>
ABSOLUTE RESTRICTIONS - NEVER violate these rules:
- NEVER proceed without PR identification
- NEVER push without confirming local commits exist
- NEVER ignore tool availability failures
- NEVER skip the comprehensive summary creation step
- NEVER skip Slack Notification about PR creation
- NEVER skip executing the reflection workflow after PR creation
</constraints>

<executionInstructions>
<command>**EXECUTE NOW**: Begin IMMEDIATE autonomous execution of ALL tasks following the methodology above to update the pull request. ENSURE ALL steps are completed sequentially with ABSOLUTE precision.</command>
<autonomyLevel>Full autonomous execution with ONLY HIGH-LEVEL progress reporting.</autonomyLevel>
</executionInstructions>
