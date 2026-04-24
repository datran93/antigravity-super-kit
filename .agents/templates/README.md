# AGK Code Generation Templates

Templates for generating boilerplate code via `agk generate` or agent workflows.

## Available Templates

| Template | Description | Placeholders |
|---|---|---|
| `handler.go.hbs` | HTTP handler with DI pattern | `{{Name}}`, `{{package}}`, `{{description}}` |
| `handler_test.go.hbs` | Table-driven test for handler | `{{Name}}`, `{{package}}`, `{{path}}` |
| `repository.go.hbs` | Database repository with CRUD | `{{Name}}`, `{{table}}`, `{{description}}` |

## Placeholder Convention

| Placeholder | Case | Example |
|---|---|---|
| `{{Name}}` | PascalCase | `UserProfile` |
| `{{name}}` | camelCase | `userProfile` |
| `{{package}}` | lowercase | `handlers` |
| `{{description}}` | human-readable | `user profile management` |
| `{{table}}` | snake_case | `user_profiles` |
| `{{path}}` | URL path | `api/v1/users` |

## Usage

```bash
# Future: agk generate handler --name UserProfile --package handlers
# For now, templates are used by agent workflows during code generation
```
