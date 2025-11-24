# AI Conformance Features (Godog)

This directory contains Gherkin `.feature` files for AI Conformance testing, executed via Godog.

## Structure

- `accelerators/`: Tests related to hardware accelerators (GPU, DRA, etc.).
- `networking/`: Tests related to AI networking (Gateway API, etc.).
- `embed.go`: Embeds the feature files into the test binary.

## Running Tests

To run these tests, pass the `--godog` flag to the test binary:

```bash
./e2e.test --godog --kubeconfig ~/.kube/config
```

## Adding New Tests

1. Create a `.feature` file in the appropriate subdirectory.
2. Implement step definitions in `../ai-steps/<category>/steps.go`.
3. Register the new steps in `../ai-steps/steps.go`.
