module github.com/RealFax/peregrine

go 1.21

require (
	github.com/gobwas/ws v1.3.0
	github.com/google/uuid v1.6.0
	github.com/jellydator/ttlcache/v3 v3.2.0
	github.com/panjf2000/ants/v2 v2.9.0
	github.com/panjf2000/gnet/v2 v2.3.6
	github.com/pkg/errors v0.9.1
)

replace github.com/gobwas/ws v1.3.0 => github.com/RealFax/ws v0.3.0

require (
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
