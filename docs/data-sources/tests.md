# circleci_tests

Retrieves test results from a CircleCI job.

## Example Usage

```hcl
data "circleci_tests" "build_tests" {
  project_slug = "gh/myorg/myrepo"
  job_number   = 123
}

# Output test summary
output "test_summary" {
  value = {
    total_tests = length(data.circleci_tests.build_tests.tests)
    passed_tests = length([
      for test in data.circleci_tests.build_tests.tests :
      test if test.result == "success"
    ])
    failed_tests = length([
      for test in data.circleci_tests.build_tests.tests :
      test if test.result == "failure"
    ])
  }
}

# Find failed tests
locals {
  failed_tests = [
    for test in data.circleci_tests.build_tests.tests :
    test if test.result == "failure"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `project_slug` - (Required) Project slug in the form `vcs-slug/org-name/repo-name`.
* `job_number` - (Required) The job number for which to retrieve test results.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `tests` - A list of test results from the job. Each test has the following attributes:
  * `message` - The test result message.
  * `source` - The test source.
  * `run_time` - The test execution time in seconds.
  * `file` - The test file path.
  * `result` - The test result (success, failure, skipped).
  * `name` - The test name.
  * `classname` - The test class name.

## Notes

* This data source is useful for analyzing test results and creating reports.
* Test results are typically collected from JUnit XML files or similar test output formats.
* The `run_time` attribute helps identify slow tests that may need optimization.
