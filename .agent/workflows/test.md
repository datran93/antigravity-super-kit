---
description: Test generation and test running command. Creates and executes tests for code.
---

# /test - Test Generation and Execution

Guide agents to generate, run, and analyze tests systematically.

---

## When to Use

- `/test` - Run all tests
- `/test [file/feature]` - Generate tests for specific target
- `/test coverage` - Show test coverage report
- `/test watch` - Run tests in watch mode

---

## Phase 1: Test Analysis 🔍

### Step 1.1: Identify Test Target

Understand what needs testing:

```markdown
### Test Target Analysis

**Target:** [file/function/feature] **Type:** Unit | Integration | E2E **Framework:** [detect from project or ask]
**Existing Tests:** [yes/no - location if yes]
```

### Step 1.2: Detect Project Test Setup

Check project for test configuration:

```markdown
### Project Test Setup

| Aspect         | Detected Value                          |
| -------------- | --------------------------------------- |
| Framework      | [Jest/Vitest/pytest/go test/etc.]       |
| Config File    | [path to config]                        |
| Test Directory | [tests/ or __tests__/ or *_test.* etc.] |
| Run Command    | [detected command]                      |
```

### Step 1.3: Identify Test Cases

```markdown
### Test Cases to Create

| Test Case                 | Type | Coverage   | Priority |
| ------------------------- | ---- | ---------- | -------- |
| Should [happy path]       | Unit | Happy path | High     |
| Should handle [edge case] | Unit | Edge case  | Medium   |
| Should reject [invalid]   | Unit | Validation | High     |
| Should handle [error]     | Unit | Error case | Medium   |
```

---

## Phase 2: Test Generation 🧪

### Step 2.1: Analyze Code Under Test

1. **Identify functions/methods** to test
2. **Find inputs and outputs** for each
3. **Detect dependencies** to mock
4. **Identify edge cases** and error conditions

### Step 2.2: Generate Test Structure

Follow the AAA pattern (Arrange-Act-Assert):

```markdown
### Test Structure

**Describe:** [Component/Module Name] **Context:** [Function/Method Name] - Test: [should do expected behavior] -
Arrange: [setup] - Act: [execute] - Assert: [verify]
```

### Step 2.3: Language-Agnostic Test Template

```
Test Suite: [Component Name]

  Test Group: [Function Name]

    Test Case: "should [expected behavior] when [condition]"
      Setup:
        - [Create test data]
        - [Configure mocks]

      Execute:
        - [Call function under test]

      Verify:
        - [Assert expected outcome]

    Test Case: "should [handle error] when [invalid input]"
      Setup:
        - [Create invalid data]

      Execute:
        - [Call function]

      Verify:
        - [Assert error is thrown/returned]
```

---

## Phase 3: Test Execution ▶️

### Step 3.1: Run Tests

Use project-specific test command:

```markdown
### Execute Tests

**Command:** [detected or configured test command] **Scope:** [all | specific file | specific test]
```

### Step 3.2: Analyze Results

```markdown
### Test Results

**Summary:**

- ✅ Passed: [count]
- ❌ Failed: [count]
- ⏭️ Skipped: [count]
- 📊 Coverage: [percentage if available]

**Failed Tests:** | Test | Error | Location | | ----------- | --------------- | ----------- | | [test name] | [error
message] | [file:line] |
```

### Step 3.3: Diagnose Failures

For each failed test:

```markdown
### Failure Analysis: [Test Name]

**Expected:** [what should happen] **Actual:** [what happened] **Likely Cause:** [hypothesis] **Suggested Fix:**
[recommendation]
```

---

## Phase 4: Coverage Analysis 📊

### Step 4.1: Generate Coverage Report

```markdown
### Coverage Report

| File/Module     | Lines | Functions | Branches |
| --------------- | ----- | --------- | -------- |
| [path/to/file]  | 85%   | 90%       | 75%      |
| [path/to/other] | 60%   | 70%       | 50%      |

**Overall:** [percentage] **Threshold:** [configured threshold] **Status:** ✅ Pass | ❌ Below threshold
```

### Step 4.2: Identify Coverage Gaps

```markdown
### Uncovered Code

| Location     | Lines       | Reason/Suggestion         |
| ------------ | ----------- | ------------------------- |
| [file:lines] | [line nums] | [add test for X scenario] |
```

---

## Phase 5: Documentation & Delivery 📝

### Step 5.1: Test Summary

```markdown
## 🧪 Test Report: [Target]

### Tests Created/Run

| Test Suite   | Passed | Failed | Skipped |
| ------------ | ------ | ------ | ------- |
| [suite name] | X      | Y      | Z       |

### Coverage

- Lines: X%
- Functions: X%
- Branches: X%

### Key Findings

- [Finding 1]
- [Finding 2]

### Recommendations

- [ ] [Add test for edge case X]
- [ ] [Increase coverage for module Y]
```

### Step 5.2: Save & Notify

1. Save to `.agent/docs/TEST-{slug}.md`
2. **Slug generation**: Extract 2-3 key words → lowercase → hyphen-separated → max 30 chars
   - "auth.service" → `TEST-auth-service.md`
   - "user registration" → `TEST-user-registration.md`
3. Notify: `✅ Test report saved: .agent/docs/TEST-{slug}.md`

---

## Quick Reference

### Workflow Flow

```
Test Analysis → Test Generation → Test Execution → Coverage Analysis → Documentation
      ↓              ↓                 ↓                ↓                  ↓
  Identify       Create tests      Run & analyze    Check coverage    Report
  targets        with AAA          results          gaps              findings
```

### Test Types

| Type            | Scope                  | When to Use            |
| --------------- | ---------------------- | ---------------------- |
| **Unit**        | Single function/method | Isolated logic testing |
| **Integration** | Multiple components    | Component interaction  |
| **E2E**         | Full user flow         | Critical user journeys |
| **Smoke**       | Basic functionality    | Quick sanity check     |

### AAA Pattern

| Phase       | Purpose                           |
| ----------- | --------------------------------- |
| **Arrange** | Set up test data and dependencies |
| **Act**     | Execute the code under test       |
| **Assert**  | Verify the expected outcome       |

---

## Key Principles

- **Test behavior, not implementation** - Focus on what, not how
- **One assertion per test** - Keep tests focused (when practical)
- **Descriptive test names** - `should [action] when [condition]`
- **Independent tests** - Tests shouldn't depend on each other
- **Mock external dependencies** - Isolate code under test
- **Fast tests** - Quick feedback loop

---

## Anti-Patterns (AVOID)

| ❌ Anti-Pattern              | ✅ Instead                       |
| ---------------------------- | -------------------------------- |
| Testing implementation       | Test behavior and outcomes       |
| Multiple assertions per test | One logical assertion per test   |
| Shared test state            | Fresh setup for each test        |
| Vague test names             | Descriptive: should X when Y     |
| Testing private methods      | Test through public interface    |
| Ignoring edge cases          | Cover boundaries and error cases |

---

## Examples

```bash
/test                           # Run all tests
/test src/services/auth         # Test auth service
/test coverage                  # Show coverage report
/test user registration flow    # Generate tests for feature
/test fix failed tests          # Analyze and fix failures
```
