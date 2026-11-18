---
name: tdd
description: Drive one behavior through red, green and refactor
argument-hint: "<behavior>"
disable-model-invocation: true
---
Implement this behavior test-first: $ARGUMENTS

1. Find the test file and suite that already cover this seam, and extend it rather than adding a low-value test.
2. Write one failing test for the smallest next slice of the behavior. Run only that package and show the red output.
3. Write the minimal code that makes it pass. Run the package again and show the green output.
4. Refactor only on green, then run the package once more.
5. Repeat from step 2 for the next slice until the behavior is complete, then run the full suite and lint.

Ask through the form if the expected behavior is ambiguous.
