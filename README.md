# sharder
A package for sharding go test suites. It uses the profile data of the test runs to determine a fair partitioning of the tests.

## Usage

```
go run . -output=output.txt -total=10 -shard=0
```

This will generate a pattern for the tests in the output file.

The output file should be a JSON file containing the test results. That can be created by running `go test -json` and redirecting the output to a file.


Example usage with GitHub Actions:

```
jobs:
  tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        shard: [0, 1, 2, 3]
  steps:
    - name: Shard tests
      run: go run github.com/VenelinMartinov/sharder --output=outputdata.json --total ${{ strategy.job-total }} --index ${{ strategy.job-index }} > tests
    - name: Run tests
      run: go test $(cat tests)
```

adapted from https://github.com/pulumi/shard