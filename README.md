Labproooo

#### Detailed Coverage
Generate detailed coverage report for all packages:
```bash
go test ./tests/... -coverprofile=coverage.out -coverpkg=./...
```

#### Function-level Coverage
View function-level coverage details:
```bash
go tool cover -func=coverage.out
```

#### HTML Coverage Report
Generate and view an HTML coverage report:
```bash
# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open in browser (Windows)
start coverage.html
```

#### Coverage Summary
Get just the total coverage percentage:
```bash
go tool cover -func=coverage.out | tail -1
```

### Example Coverage Commands

```bash
# Run tests with coverage and generate HTML report
go test ./tests/... -coverprofile=coverage.out -coverpkg=./...
go tool cover -html=coverage.out -o coverage.html

# View coverage summary
go tool cover -func=coverage.out | grep "total:"
```
