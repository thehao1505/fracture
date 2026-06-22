I want to build a login page for this app. Some requirements are:
- Use describle_images tool to understand the mockup designs at .github/mockups/login/folder. Then implement the login page according to the design.
    - .github/mockups/login/login-form.png: for general design
    - .github/mockups/login/login-failed.png: for log in failed follow this
    - .github/mockups/login/login-with-validation.png: for validation messages.
    - Password field should have eye icon to toggle show/hide password
    - The loging page must be central of the page
- Any Navigation to the page should folllow these rules: 
    - if token is available, redirect to "/dashboard"
    - if token is not available, redirect to "/login"
- Define login page in this src/app/configuration-ui/login/
- Create Auth service to retrieve login credential from the login form and call this backend API to exchange token
```
    POST https://llm-proxy.kms-technology.com/configuration/token
    Content-Type: application/json
    User-Agent: PostmanRuntime/7.46.1
    Accept: */*
    Postman-Token: 7cd7bc5e-fe1a-4a35-955b-bcb8a8fe0ade
    Host: llm-proxy.kms-technology.com
    Accept-Encoding: gzip, deflate, br
    Connection: keep-alive
    Content-Length: 108
    
    {
    "clientId": "xxx",
    "clientSecret": "xxx"
    }
```
- After user provided credential, the login button change from disabled to enabled
- Once user click the login button, the button should show loading state and send request Post to the Auth service for login.
- If success, responsed token should be stored in cookie(best secure this) for further use and redirect to "/dashboard"
- If failed, show error message like in the design