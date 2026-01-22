# Planning Guidelines for Agents

## Overview

This document outlines the rules and guidelines for creating and maintaining implementation plans in the `plans/` directory.

## Mandatory Rules for Planning

### Rule 1: Always Create a Progress File

For every new implementation plan, you **MUST** create a corresponding progress tracking file:

- **Naming Convention**: `PROGRESS_{PLAN_NAME}.md` (e.g., `PROGRESS_IMPORT_API_COLLECTIONS.md`)
- **Location**: Same directory as the plan (typically `plans/`)
- **Purpose**: Track step completion, document decisions, and enable handoff between agents

**Progress File Structure**:
```markdown
# Progress: {Plan Name}

This file tracks progress through the {PLAN_FILE}.md implementation plan. **After each completed step, append an entry below.**

## Progress Entries

### Entry Template
```markdown
## [YYYY-MM-DD HH:MM] - Step Name

**Completed:**
- What was done

**Files Changed:**
- path/to/file1.go
- path/to/file2.ts

**How to Verify:**
- Command to run or test to execute

**Next Steps:**
- What comes next

**Blockers/Notes:**
- Any issues or decisions made
```

---

## [YYYY-MM-DD HH:MM] - Step Name

**Completed:**
- Bullet points of what was accomplished

**Files Changed:**
- List of files created/modified

**How to Verify:**
- Specific commands or tests to run

**Next Steps:**
- What comes next in the plan

**Blockers/Notes:**
- Any blockers, decisions, or notes
```

### Rule 2: Each Step Must Be Detailed

Every step in an implementation plan **MUST** include:

1. **Goal Statement**: Clear description of what this step accomplishes
2. **Files**: List of files to create and/or modify
3. **Output**: What this step produces
4. **Detailed Tasks**: Numbered subtasks with code examples where appropriate
5. **Testing**: How to verify this step works

**Step Template**:
```markdown
### Step N: Step Title

**Goal**: One sentence describing what this step accomplishes

**Files**:
- Create: `path/to/new/file.go`
- Modify: `path/to/existing/file.go`

**Output**: What this step produces (e.g., "MultiWorkspaceStore implementation")

**Detailed Tasks**:
N.1. First subtask with implementation details
N.2. Second subtask with code examples
N.3. Third subtask...

**Testing**:
- Test description for this step
```

## Best Practices

### Keep Plans Actionable
- Each step should be completable in one session
- Avoid overly large steps - break them down
- Each step should produce verifiable output

### Document Decisions
- When making architectural choices, document the reasoning
- Note alternatives considered and why they were rejected
- Record any constraints or requirements discovered

### Update Progress
- Append to PROGRESS file immediately after completing a step
- Include files changed for easy review
- Note any blockers or decisions made

### Handle Handoff
- PROGRESS file should enable another agent to pick up where you left off
- Include "Next Steps" in each entry
- Note any incomplete work or partial implementations

## Plan File Structure

All plans should be saved in `constrictor-rest-client/plans/` with:

1. `{PLAN_NAME}.md` - Main implementation plan
2. `PROGRESS_{PLAN_NAME}.md` - Progress tracking file
3. `agents.md` - This file (shared across plans)

## Example Workflow

1. **Create Plan**: Write detailed plan to `{PLAN_NAME}.md`
2. **Create Progress File**: Create `PROGRESS_{PLAN_NAME}.md` with plan entry
3. **Execute Step**: Implement step from plan
4. **Document Progress**: Append completed step entry to progress file
5. **Repeat**: Continue with next step until complete

## Review Checklist

Before considering a plan complete:

- [ ] Plan file exists with detailed steps
- [ ] Progress file exists with initial entry
- [ ] Each step has goal, files, output, detailed tasks, and testing
- [ ] Steps are actionable and independently verifiable
- [ ] Dependencies between steps are documented
