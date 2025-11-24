# AI Conformance Features (Godog)

This directory contains Gherkin `.feature` files for AI Conformance testing, executed via Godog.

## Structure

- `accelerators/`: Tests related to hardware accelerators (GPU, DRA, etc.).
- `networking/`: Tests related to AI networking (Gateway API, etc.).
- `security/`: Tests related to secure accelerator access (Device Plugin, etc.).
- `orchestration/`: Tests related to workload scheduling and scaling (Kueue, Autoscaling).
- `observability/`: Tests related to monitoring and metrics (Prometheus).
- `embed.go`: Embeds the feature files into the test binary.

## Running Tests

To run these tests, pass the `--features` flag to the test binary:

```bash
./e2e.test --features --kubeconfig ~/.kube/config
```

## Adding New Tests

1. Create a `.feature` file in the appropriate subdirectory.
2. Implement step definitions in `../ai-steps/<category>/steps.go`.
3. Register the new steps in `../ai-steps/steps.go`.