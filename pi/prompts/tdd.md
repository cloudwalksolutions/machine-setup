---
description: Drive one behavior through red → green → refactor
argument-hint: "<behavior>"
---
Implement this behavior with strict TDD: $@

1. Find the test file and suite that already cover this seam; extend it rather than adding a low-value test.
2. Write ONE failing test for the smallest next slice of the behavior. Run only that package and show me the red output.
3. Write the minimal code that makes it pass. Run the package again and show the green output.
4. Refactor only on green, then run the package once more.
5. Repeat from step 2 for the next slice until the behavior is complete, then run the full suite and lint.

Do not write source before its failing test. Do not batch tests. Stop and ask if the expected behavior is ambiguous.
