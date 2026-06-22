# Workflow Optimization Notes:

1. This workflow is primarily built for coding convention review. You might want to introduce other types of reviews by feeding different documents and git filters. For example: FE style guide review, security review, performance review. By cloning the workflow and providing the specific document and filter for each review type.
2. You might want to change the commit message and cadence that works with your workflow. Note that the more granular commits you have, the more control you might have.

# Required Changes:

1. The coding convention documents of your project
2. Update the git diff filter to match your project language and code review purposes. You might want to introduce more filters to have the agent review a specific area of the project with more specific requirements.
E.g:
For TS project
git --no-pager diff --name-only @{u}..HEAD -- '*.ts*' ':!*.spec.ts'

For C# project
git --no-pager diff --name-only @{u}..HEAD -- '*.cs*' ':!*.spec.cs'

# Usage Tips:

- Keep the review document short with bullet points and examples
- Different types of reviews should be different workflows.
- Use git diff filters to focus on specific files or directories relevant to the review.
