# Workflow Optimization Notes:

1. Before doing testing, run the a.II.Onboarding-Testing.prompt.md workflow and make sure the generated knowledge captures your project's testing conventions and design.
2. Specify the testing file extensions that you want to perform unit tests on after the implementation step.
   `git --no-pager diff --name-only @{u}..HEAD -- '*controller.ts*' '*service.ts*' '*provider.ts*' '*repository.ts*'`
3. Increase the -U100 parameter for more context in the diff output if your project has very long functions. U100 means get the top and bottom 100 lines of the changed line.
   `git --no-pager diff @{u}..HEAD -U100 -- <file_path>`
4. You might want to change the commit message and cadence that works with your workflow. Note that the more granular commits you have, the more control you might have.

# Required Changes:

1. Change the git diff filter to your project language with a filter on the files you want to perform unit tests on.
2. Provide testing convention documents of your project in e.implementation-testing-quality.reflection.md
3. Update the git diff filter in e.implementation-testing-quality.reflection.md to match your project language and testing files.
   E.g:
   For TS project
   `git diff --name-only HEAD -- '*.test.ts*' '*spec.ts*'`

For C# project
`git --no-pager diff --name-only @{u}..HEAD -- '*.cs*' ':!*.spec.cs'`

# Usage Tips:

- Provide more specific filters in the plan to help the agent focus on the right files to perform testing
- If the number of tests is very large, you might want to break the testing plan document into smaller test cases to execute per session.
