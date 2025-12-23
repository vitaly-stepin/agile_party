## Collaboration & Coding Guidelines

1. Clarification First
If any part of my request is ambiguous, incomplete, or could reasonably be interpreted in multiple ways, you must ask clarifying questions before proceeding. Do not make assumptions.

2. Critical Evaluation of Suggestions
You are expected to reason critically about my suggestions:
- Apply them only if they are technically sound and beneficial.
- If a better or more robust alternative exists, propose it clearly and explain why.
- Do not follow instructions blindly; prioritize correctness and maintainability.

3. Tech Stack & Idioms
- Backend: Go - follow idiomatic Go patterns, use standard library when sufficient
- Frontend: React/TypeScript - prefer functional components and hooks
- Write code that fits naturally within the existing architecture

4. Performance & Optimization
- Write performance-conscious code with attention to latency
- Avoid premature optimization, but eliminate obvious bottlenecks
- Explicitly call out trade-offs when optimizing
- In Go: be mindful of allocations, use goroutines judiciously for concurrent operations

5. Error Handling
- Go: always handle errors explicitly, wrap errors with context using fmt.Errorf or errors.Wrap
- Frontend: handle error states in UI, provide meaningful user feedback

6. Testing Philosophy
- Write tests for non-trivial business logic and critical paths
- Prioritize integration tests for core workflows
- Unit tests for complex algorithms or edge cases

7. Production Quality
- Default to production-ready code
- Highlight edge cases, failure modes, and scalability concerns
- Consider security implications (input validation, authentication, etc.)

8. Code Style
- Add comments only where absolutely necessary, skip obvious comments
- Aim for self-documenting code with clear naming
- Keep functions focused and composable
