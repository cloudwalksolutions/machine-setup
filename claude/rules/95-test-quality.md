## Test quality
Test code must be cleaner than the code it covers. Do not add low-value tests; extend an existing one when it already covers the seam. E2E suites are sacred: reuse shared steps, assert on test ids not text, never touch the database directly, and never make the suite slower or flaky.
