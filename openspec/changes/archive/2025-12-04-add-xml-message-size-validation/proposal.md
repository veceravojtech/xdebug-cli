# Add XML Message Size Validation

## Why

When debugging code with deep stack traces (e.g., views/templates with 600+ line backtraces), the DBGp message parsing fails with:

```
invalid message size 'php\" lineno=\"622\">...</init>': strconv.Atoi: parsing "...": invalid syntax
```

This happens because:
1. Large Xdebug responses may have corrupted framing or split across TCP packets
2. The size field is parsed without bounds checking - corrupted values could cause memory exhaustion
3. No validation that size field contains only digits before `strconv.Atoi()`

## What Changes

- Add maximum message size constant (e.g., 100MB)
- Validate size field format before parsing (digits only)
- Add bounds checking after parsing to prevent memory exhaustion
- Improve error messages to aid debugging

## Impact

- Affected specs: `dbgp` (Connection Message Framing requirement)
- Affected code: `internal/dbgp/connection.go:26-59`
- No breaking changes - stricter validation only
- Prevents crashes on corrupted Xdebug responses
