# SDLC Workflow Apply Guide

1. Copy all file and document in this folder into .github/
2. Provide project's documents e.g Coding/Testing convention, PR creation/update templates..
3. Create toolset file following the content of workflow.toolsets.jsonc
4. Follow \*.optimize.md documents to shape the workflow into the project's context.

# Lookup SLACK_MCP_XOXC_TOKEN

1. Open Slack Web App and Press F12 to open your browser's Developer Console.
2. In Firefox, under Tools -> Browser Tools -> Web Developer tools in the menu bar
3. In Chrome, click the "three dots" button to the right of the URL Bar, then select More Tools -> Developer Tools
4. Switch to the console tab.
5. Type "allow pasting" and press ENTER.
6. Paste the following snippet and press ENTER to execute: JSON.parse(localStorage.localConfig_v2).teams[document.location.pathname.match(/^\/client\/([A-Z0-9]+)/)[1]].token
