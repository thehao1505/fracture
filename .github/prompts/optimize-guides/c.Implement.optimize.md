# Workflow Optimization Notes:

1. While working through the implementation steps, the agent will make some coding mistakes where it does not follow the project conventions or best practices. Hence, the c.implementation-coding-quality.reflection workflow will perform a self-review at the end of the implementation session to make sure the agent follows project conventions. You might want to introduce bullet points with examples for all the conventions that you expect the agent to follow.
2. If there are any other documents that apply to your coding flow, things like style guides, utility utilization, or anything special to your project, make sure to reference them as well.
3. Keep those convention files minimal. You might want to use the code review flow for the agent to perform the self-review step through a more extensive list of rules and conventions.

# Require Changes:

1. Review c.implementation-coding-quality.reflection.md and change the git diff command to filter the files that you want to refactor. In this workflow, it performs `git diff --name-only HEAD -- '*.ts*' ':!*.spec.ts'` to detect all the project coding files (excluding test files). Therefore, you might want to change the filter so it matches your project language.
2. You might want to revise the commit message convention to match your project's practice.

E.g:
For TS project
git --no-pager diff --name-only @{u}..HEAD -- '*.ts*' ':!*.spec.ts'

For C# project
git --no-pager diff --name-only @{u}..HEAD -- '*.cs*' ':!*.spec.cs'

# Usage Tips:

- A pre-commit hook is not just nice to have, but a must-have to ensure agent code quality and consistency.
