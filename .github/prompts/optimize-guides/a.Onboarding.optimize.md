# Workflow Optimization Notes:

1. Review the onboarding scanning phases to understand how this workflow performs the project scanning for the onboarding process for both coding and testing-related tasks.
2. For each task in the phase, you might want to provide example folders that are specific to your project structure and conventions so the agent can reference them to produce better scanning results.
3. You might want to introduce a new phase/task to cover any additional aspects of the project that you think would be beneficial for the agent while coding. Think about patterns, conventions, and project structure that should be included.

# Usage Tips:

1. Make sure your project doesn't have logs, dist, or any temporary files that are not in .gitignore. The agent will see and scan them, which is a waste.
2. If your project is too large, scan the specific projects/services/folders that you are going to work on. This helps narrow down the context and improve the relevance of the results.
