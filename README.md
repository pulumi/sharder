# sharder
A package for sharding go test suites. It uses the profile data of the test runs to determine a fair partitioning of the tests. It will generate a pattern to pass to `go test` to run the tests in each shard.

```
go run . -output=test-timings.jsonl -total=3 -index=0
```

This will generate a pattern to pass to `go test` to run the tests in the shard.

For example for the tests `test1`, `test2`, `test3` of equal duration, the output will be:
- `-run \"(^test1$)\"` for the first shard
- `-run \"(^test2$)\"` for the second shard
- `-skip \"(^test1$)|(^test2$)\"` for the third shard

Note the final shard generates a skip pattern and it will catch any tests which are not in the timings file. This avoids the need to regenrate the timings each time a test is added. This means that if `test4` is added but the timings are not updated, it will still be run in the final shard.

## Usage

Generate a JSON file containing the test timings:

```
go test ./... -json > test-timings.jsonl
```

Generate the shard pattern:

```
go run . -output=test-timings.jsonl -total=3 -index=0
```

Example usage with GitHub Actions:

```
jobs:
  tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        shard: [0, 1, 2]
  steps:
    - name: Shard tests
      run: go run github.com/pulumi/sharder --output=test-timings.jsonl --total ${{ strategy.job-total }} --index ${{ strategy.job-index }} > tests
    - name: Run tests
      run: go test $(cat tests)
```

Inspired by https://github.com/pulumi/shard. Note that `shard` is different in the following ways:
- `shard` does not require any prior runs of the tests. It will parse the test files and generate patterns based on the test names.
- `shard` does not have knowledge of test timings, so can not generate as fair a partition as `sharder`.
