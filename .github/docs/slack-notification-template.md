# Slack Notification Template
This template is used to send notifications to a designated Slack channel about Pull Request (PR) activities, such as creation or updates. 
The notification includes essential details like the ticket ID, PR link, and a summary of changes.

## Slack Channel ID 
Use this Slack Channel ID to post the below message: **C09F6G9P350**

## Message Content
### Ticket ID instruction
The Ticket ID should be extracted from the most recent commit messages that follows the established commit convention.
Execute `git log --oneline -3` to Identify the Ticket ID from the last three commit messages. Convention example: `TICKET-ID | <description>`.

### Slack Notification Format
Fill dynamic data in [ brackets ] as indicated below:

```md
[Agent]: PR Notification
PR [CREATED/UPDATED] for Ticket Id: [TICKET-ID]
PR Link: [PR Link]
Summary of Changes: [Bullet points from the comprehensive summary]
```
