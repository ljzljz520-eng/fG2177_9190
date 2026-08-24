# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
--- FAIL: TestMissingSubmissionReturnsFriendlyResult (0.00s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered, repanicked]
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x17c688]

goroutine 22 [running]:
testing.tRunner.func1.2({0x1b2080, 0x379200})
	/usr/local/go/src/testing/testing.go:1974 +0x1a0
testing.tRunner.func1()
	/usr/local/go/src/testing/testing.go:1977 +0x318
panic({0x1b2080?, 0x379200?})
	/usr/local/go/src/runtime/panic.go:860 +0x12c
courseworkledger/internal/service.(*Catalog).BuildDashboard(0x40d082e02270)
	/app/internal/service/dashboard.go:62 +0x6b8
courseworkledger.TestMissingSubmissionReturnsFriendlyResult(0x40d082e65208)
	/app/regression_test.go:205 +0x2c0
testing.tRunner(0x40d082e65208, 0x1f8b30)
	/usr/local/go/src/testing/testing.go:2036 +0xc4
created by testing.(*T).Run in goroutine 1
	/usr/local/go/src/testing/testing.go:2101 +0x3a8
FAIL	courseworkledger	0.022s
?   	courseworkledger/cmd/coursework	[no test files]
ok  	courseworkledger/internal/domain	0.001s
ok  	courseworkledger/internal/importer	0.001s
ok  	courseworkledger/internal/policy	0.001s
ok  	courseworkledger/internal/report	0.001s
ok  	courseworkledger/internal/service	0.007s
ok  	courseworkledger/internal/store	0.006s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/coursework): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/coursework): exit `0`
