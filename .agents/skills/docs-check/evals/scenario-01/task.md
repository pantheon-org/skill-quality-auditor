# Scenario 01: Basic Build Check

## User Prompt

"Check the docs build — make sure the site compiles and tell me how many pages it generated."

## Expected Behavior

1. Run `npx @docmd/core build` from the repo root.
2. Wait for the build to complete.
3. Verify the exit code is 0.
4. Extract the page count from the build output ("Generated N pages").
5. Confirm the `site/` output directory was created.
6. Report the result: build passed, page count, and output directory location.
7. If the build fails, read the error output, identify the cause, and suggest a fix.

## Success Criteria

- `npx @docmd/core build` is executed.
- Exit code is verified as 0.
- Page count is extracted and reported.
- `site/` directory existence is confirmed.
- User receives a clear pass/fail summary.

## Failure Conditions

- Build command is not run.
- Exit code is not checked.
- Page count is not reported.
- Build failure is not diagnosed.
- User receives only a raw command dump without interpretation.

**Context:**

- Repository root: current working directory
- `docmd.config.json` exists at repo root with `"src": "./docs", "out": "./site"`
- Pre-requisite: Node.js 18+ available via `npx`
