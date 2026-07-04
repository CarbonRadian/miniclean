# miniclean

A small, dependency-free CLI for cleaning structured data in CSV and JSON files.

`miniclean` reads tabular data, applies a configurable pipeline of cleaning
rules, and writes the result back out — suitable for quick fixes in shell
pipelines or as a preprocessing step in larger data workflows.

## Status

Work in progress. See the roadmap below.

## Planned features

- [ ] Core cleaning rules: trim whitespace, drop empty rows, deduplicate rows,
      normalize headers, normalize null-ish values (`NA`, `N/A`, `null`, ...)
- [ ] CSV input/output
- [ ] JSON input/output (arrays of objects)
- [ ] Reads stdin / writes stdout for pipeline use
- [ ] Cross-format conversion (CSV → JSON and back)

## License

MIT
