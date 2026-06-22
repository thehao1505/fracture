# Project Testing Practices

1. **Test Location and Naming**  
   All test files must be placed inside the `./test` folder, located at the same directory level as the files they test.  
   **Example:**

   ```
   src/
     users/
       users.service.ts
       test/
         users.service.spec.ts
   ```

   The test file is named as `<original-file-name>.spec.ts`.

2. **Testing Utility Functions**  
   All reusable testing utilities and helper functions must be placed under:

   ```
   libs/common/src/test-utils/
   ```

   Utilities here can be imported in any test file as needed.

3. **Testing Framework & Best Practices**  
   Use NestJS testing best practices, including:
   - Use the Nest testing module (`Test.createTestingModule`) for unit & integration tests
   - Mock dependencies properly
   - Cover positive, negative, and edge cases
   - Use async/await patterns correctly for asynchronous tests
   - Follow proper test isolation and cleanup

4. **Testing Approaches**

- Follow the Arrange-Act-Assert convention for tests.
- Name test variables clearly.
  - Follow the convention: inputX, mockX, actualX, expectedX, etc.
- Write unit tests for each public function.
  - Use test doubles to simulate dependencies.
    - Except for third-party dependencies that are not expensive to execute.
- Write acceptance tests for each module.
  - Follow the Given-When-Then convention.
