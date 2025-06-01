module github.com/bresrch/sawmill/benchmark

go 1.24.1

replace github.com/bresrch/sawmill => ../

require (
	github.com/bresrch/sawmill v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
	go.uber.org/zap v1.27.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
