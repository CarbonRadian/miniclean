# miniclean

A small, dependency-free CLI for cleaning structured data in CSV and JSON files.

`miniclean` reads tabular data, applies a configurable pipeline of cleaning
rules, and writes the result back out — suitable for quick fixes in shell
pipelines or as a preprocessing step in larger data workflows.

## Install

```sh
go install github.com/CarbonRadian/miniclean/cmd/miniclean@latest
```

Or build from source:

```sh
git clone https://github.com/CarbonRadian/miniclean.git
cd miniclean
go build -o miniclean ./cmd/miniclean
```

## Usage

```sh
miniclean [flags] [input-file]
```

Reads the input file (or stdin), applies the selected cleaning rules, and
writes the result to the output file (or stdout).

### Flags

| Flag                | Description                                                          |
| -------------------- | --------------------------------------------------------------------- |
| `-o file`            | Output file (default stdout)                                        |
| `-from csv\|json`    | Input format (default: inferred from extension, `csv` on stdin)     |
| `-to csv\|json`      | Output format (default: same as input format)                       |
| `-trim`              | Trim whitespace from headers and cells                              |
| `-normalize-nulls`   | Replace null-ish values (`NA`, `N/A`, `null`, `none`, `nil`, `-`) with empty |
| `-drop-empty`        | Drop rows whose cells are all empty                                 |
| `-dedup`             | Drop exact duplicate rows, keeping the first                        |
| `-normalize-headers` | Lowercase headers and convert to `snake_case`                       |
| `-version`           | Print version and exit                                              |

With no rule flags, **all rules are applied**. Pass one or more rule flags to
run only those rules.

### Examples

Clean a CSV file in place with all rules, writing to stdout:

```sh
miniclean data.csv
```

Only trim whitespace and drop empty rows, writing to a new file:

```sh
miniclean -trim -drop-empty -o clean.csv data.csv
```

Convert CSV to JSON:

```sh
miniclean -to json data.csv > data.json
```

Convert JSON to CSV, applying every rule:

```sh
miniclean -from json -to csv data.json > data.csv
```

Works in a shell pipeline via stdin/stdout:

```sh
cat data.csv | miniclean -dedup > deduped.csv
```

## Cleaning rules

- **trim** — removes leading/trailing whitespace from every header and cell
- **normalize-nulls** — replaces common null placeholders (`NA`, `N/A`, `null`,
  `none`, `nil`, `-`, case-insensitive) with the empty string
- **drop-empty** — removes rows where every cell is empty
- **dedup** — removes exact duplicate rows, keeping the first occurrence
- **normalize-headers** — lowercases headers and converts spaces/dashes to
  underscores (e.g. `First Name` → `first_name`)

## Development

```sh
go build ./...
go test ./...
go vet ./...
```

Or via the Makefile: `make build`, `make test`, `make vet`, `make lint`.

## License

MIT
