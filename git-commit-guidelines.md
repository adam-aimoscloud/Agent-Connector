# Git Commit Message Guidelines

## Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

## Types
- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that do not affect the meaning of the code (white-space, formatting, etc)
- **refactor**: A code change that neither fixes a bug nor adds a feature
- **perf**: A code change that improves performance
- **test**: Adding missing tests or correcting existing tests
- **chore**: Changes to the build process or auxiliary tools and libraries

## Scope
The scope should be the name of the component affected (as perceived by the person reading the changelog).

Examples:
- **auth**: Authentication related changes
- **api**: API related changes
- **ui**: User interface changes
- **database**: Database related changes
- **config**: Configuration changes

## Subject
The subject contains a succinct description of the change:
- Use the imperative, present tense: "change" not "changed" nor "changes"
- Don't capitalize the first letter
- No dot (.) at the end

## Body
Just as in the subject, use the imperative, present tense. The body should include the motivation for the change and contrast this with previous behavior.

## Footer
The footer should contain any information about Breaking Changes and is also the place to reference GitHub issues that this commit closes.

## Examples

### Simple commit
```
feat(auth): add JWT token authentication
```

### Commit with body
```
fix(api): resolve user login validation issue

The validation was failing for users with special characters in their email.
Updated the regex pattern to properly handle all valid email formats.

Fixes #123
```

### Breaking change
```
feat(api): change user authentication endpoint

BREAKING CHANGE: The /auth endpoint now requires a different payload structure.
Old: { username, password }
New: { email, password, rememberMe }
```

### Multiple changes
```
feat(dashboard): add user management interface

- Add user list with pagination
- Implement user creation form
- Add user edit and delete functionality
- Include role-based access control

Closes #45, #67, #89
```

## Common Commit Types for This Project

### Frontend Changes
```
feat(ui): add new dashboard component
fix(form): resolve validation error handling
style(layout): improve responsive design
refactor(components): extract reusable modal component
```

### Backend Changes
```
feat(api): implement rate limiting middleware
fix(auth): resolve JWT token expiration issue
perf(database): optimize user query performance
test(service): add unit tests for user service
```

### Configuration Changes
```
chore(config): update environment variables
docs(readme): update installation instructions
chore(deps): update dependencies to latest versions
```

### Documentation Changes
```
docs(api): add endpoint documentation
docs(readme): update project setup guide
docs(config): add configuration examples
``` 