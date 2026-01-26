# Development Constraints

## Code Changes Philosophy

### Avoid Over-Engineering
- Make only changes directly requested or clearly necessary
- Keep solutions simple and focused
- Don't add features, refactor code, or make "improvements" beyond what's asked
- A bug fix doesn't need surrounding code cleaned up
- A simple feature doesn't need extra configurability

### Error Handling & Validation
- Don't add error handling, fallbacks, or validation for scenarios that can't happen
- Trust internal code and framework guarantees
- Only validate at system boundaries (user input, external APIs)
- Don't use backwards-compatibility shims when you can just change the code

### Abstractions & Complexity
- Don't create helpers, utilities, or abstractions for one-time operations
- Don't design for hypothetical future requirements
- Keep complexity to the minimum needed for the current task
- Reuse existing abstractions where possible
- Follow the DRY principle

### Git & Build Management
- Don't create git diff noise - keep diffs clean for easy review
- Don't ask to execute builds unless explicitly requested by user

### Measurement
- Use `go test -bench` and/or `benchstat` or equivalent benchmark tools to verify improvements
- Profile with `pprof` before optimizing blind
