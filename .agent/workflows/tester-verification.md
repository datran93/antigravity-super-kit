---
description: Structured workflow for Testing. Handles test generation, execution, and verification of implemented code.
---

# 🧪 Tester / Verification Workflow

This workflow guides you through verifying the functionality and stability of code changes. It focuses on comprehensive coverage, edge-case validation, and ensuring the project remains bug-free.

## 🚀 Verification Phase

### Phase 1: Intake & Strategy 📥
Load the context of the implemented code.
- Understand the implementation logic that needs testing.
- Actively use `@mcp:skill-router` (`search_skills`) to find the exact test execution and test coverage commands for the relevant framework (e.g., `go test -cover`, `pytest --cov`, Jest). Do not guess these commands.

### Phase 2: Test Engineering 🧪
Create comprehensive test cases to challenge the implementation.
- **Unit Tests**: Write surgical tests for new functions/logic.
- **Integration Tests**: Verify components interact correctly with services (Database, Redis, etc.). All integration tests must be written in the `./test` folder.
- **Edge Cases**: Target boundary conditions and error paths.
- Use `write_to_file` to create/update test files in the appropriate test directory.

### Phase 3: Execution & Analysis ▶️
Run the test suite and evaluate the results.
- Execute tests using `run_command` (e.g., `npm test`, `go test ./...`, `pytest`).
- **MANDATORY**: Capture and analyze the logs for any failures.

### Phase 4: Feedback & Resolution 📝
Act on the test results.
- **PASS**: If all tests pass, and coverage is >= 70%, summarize the coverage and confirm stability. Transfer to `planner` role and mark the overall step as complete via `@mcp:context-manager` (`complete_task_step`), passing in the `active_files` array.
- **FAIL**: Analyze the error logs. If the failure is due to a bug in the implementation code, transition back to the `coder` role to fix it. If the failure is due to a flaw in the test code itself, fix the test code yourself in this `tester` role.
- **Circuit Breaker**: If you reach 3 consecutive FAIL cycles for the same logic, ABORT the `coder` ↔ `tester` loop to prevent token wastage. Transition to the `planner` role to re-evaluate the architecture, or the `project-manager` role to ask the USER for guidance.

### Phase 5: Role Transition 🔄
- Once tests pass, transition back to the `planner` role to pick up the next task in the plan.

## 🔴 Critical Constraints
1. **Automated Verification**: Always run actual commands; never assume code works based on a visual scan.
2. **No Implementation Fixes**: You MUST NOT modify the implementation code directly. You are only allowed to write or modify files in the test directories.
3. **Minimum Test Coverage**: It is MANDATORY to ensure that test coverage is equal to or greater than 70%. If it is below 70%, write more tests. ALWAYS print out the coverage percentage using proper framework tooling.
4. **Role Anchoring**: ALWAYS prefix every conversational response with `[Role: 🧪 Tester]`.

---

> [!IMPORTANT]
> Use the exact error logs to ensure you fix the root cause quickly when you switch back to the Coder role.
