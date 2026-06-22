# Workflow Optimization Notes:

1. This workflow detects unresolved PR feedback via conversation. So you might want to resolve the conversation by explicit comment instead of hitting the resolve button. GitHub MCP doesn't return PR comment status so the agent can't detect whether a comment is resolved or not.
2. You might want to change the commit message and cadence that works with your workflow. Note that the more granular commits you have, the more control you might have.
3. You might want to change the commit message and cadence that works with your workflow. Note that the more granular commits you have, the more control you might have.

# Required Changes:

1. Change the git diff filter in g.coding-quality.reflection.md so it matches your coding and testing project language

# Usage Tips:

1. If the PR feedback is very large, you might want to break the workflow into 3 steps: Pull conversation > Solution for each feedback > Address feedback.
2. You can optionally mark which comments the agent should handle and which ones should be ignored so you can handle them.
