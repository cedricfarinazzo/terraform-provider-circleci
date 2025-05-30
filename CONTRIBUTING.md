# Contributing to Terraform Provider for CircleCI

We welcome contributions to the Terraform Provider for CircleCI! This document provides guidelines for contributing to this project.

## Development Setup

### Prerequisites

- [Go](https://golang.org/doc/install) >= 1.21
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- CircleCI API token for testing

### Environment Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/cedricfarinazzo/terraform-provider-circleci
   cd terraform-provider-circleci
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   ```bash
   export CIRCLECI_TOKEN="your-api-token-here"
   ```

### Building the Provider

```bash
go build
```

### Running Tests

#### Unit Tests
```bash
go test ./...
```

#### Acceptance Tests
```bash
export TF_ACC=1
export CIRCLECI_TOKEN="your-test-token"
go test ./... -v
```

**Note:** Acceptance tests run against the actual CircleCI API and may create/modify real resources.

## Project Structure

```
├── main.go                     # Provider entry point
├── internal/provider/          # Provider implementation
│   ├── provider.go            # Main provider configuration
│   ├── client.go              # CircleCI API client
│   ├── resource_*.go          # Resource implementations
│   ├── data_source_*.go       # Data source implementations
│   └── *_test.go              # Test files
├── docs/                      # Documentation
│   ├── resources/             # Resource documentation
│   └── data-sources/          # Data source documentation
└── examples/                  # Usage examples
```

## Contributing Guidelines

### Adding New Resources

1. Create the resource file: `internal/provider/resource_<name>.go`
2. Implement the required interfaces:
   - `Metadata()` - Define resource name
   - `Schema()` - Define resource schema
   - `Create()`, `Read()`, `Update()`, `Delete()` - CRUD operations
3. Register the resource in `internal/provider/provider.go`
4. Add tests: `internal/provider/resource_<name>_test.go`
5. Add documentation: `docs/resources/<name>.md`
6. Add examples if appropriate

### Adding New Data Sources

1. Create the data source file: `internal/provider/data_source_<name>.go`
2. Implement the required interfaces:
   - `Metadata()` - Define data source name
   - `Schema()` - Define data source schema
   - `Read()` - Data retrieval logic
3. Register the data source in `internal/provider/provider.go`
4. Add tests: `internal/provider/data_source_<name>_test.go`
5. Add documentation: `docs/data-sources/<name>.md`
6. Add examples if appropriate

### Code Style

- Follow Go conventions and use `gofmt`
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and single-purpose
- Handle errors appropriately

### Testing

- Write unit tests for all new functionality
- Include both positive and negative test cases
- Test edge cases and error conditions
- Use the existing test patterns as reference
- Ensure acceptance tests clean up after themselves

### Documentation

- Document all resources and data sources
- Include usage examples
- Document all attributes and their types
- Include import instructions where applicable
- Keep documentation up to date with code changes

## Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch from `main`
3. **Make** your changes following the guidelines above
4. **Add** or update tests as needed
5. **Update** documentation
6. **Run** tests to ensure everything works
7. **Commit** your changes with descriptive messages
8. **Push** to your fork
9. **Create** a pull request

### Pull Request Requirements

- [ ] All tests pass
- [ ] New functionality includes tests
- [ ] Documentation is updated
- [ ] Code follows project conventions
- [ ] Commit messages are descriptive
- [ ] No breaking changes (or clearly documented)

## API Guidelines

### CircleCI API Integration

- Follow CircleCI API v2 patterns
- Handle API rate limiting gracefully
- Use appropriate HTTP methods
- Include proper error handling
- Support pagination where needed
- Respect API response structures

### Error Handling

- Use Terraform's diagnostic system
- Provide clear, actionable error messages
- Include relevant context in errors
- Handle API errors appropriately
- Don't expose sensitive information in errors

### Resource Lifecycle

- Implement proper state management
- Handle resource dependencies
- Support import functionality where possible
- Manage sensitive attributes correctly
- Handle resource updates appropriately

## CircleCI API Resources Covered

### Current Resources (14)
- Contexts and Environment Variables
- Projects and Checkout Keys
- Webhooks and Schedules
- Pipelines and Jobs
- OIDC Tokens and Policies
- Users and Usage Exports
- **Runners and Runner Tokens** (New)

### Current Data Sources (10)
- Context, Project, Insights
- Organization, Workflow, Workflows
- Policies
- **Artifacts, Tests, Jobs** (New)

## Getting Help

- Check existing [issues](https://github.com/cedricfarinazzo/terraform-provider-circleci/issues)
- Review [CircleCI API documentation](https://circleci.com/docs/api/v2/)
- Look at existing code for patterns
- Ask questions in issues or discussions

## Code of Conduct

Please be respectful and professional in all interactions. We're here to build great tools together!

## License

By contributing, you agree that your contributions will be licensed under the Mozilla Public License 2.0.
