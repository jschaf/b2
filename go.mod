module github.com/jschaf/b2

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/alecthomas/chroma v0.7.1
	github.com/cloudflare/tableflip v1.0.0
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/evanw/esbuild v0.7.16
	github.com/fsnotify/fsnotify v1.4.7
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/google/go-cmp v0.5.2
	github.com/gorilla/websocket v1.4.1
	github.com/graemephi/goldmark-qjs-katex v0.3.0
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jschaf/bibtex v0.0.0-20200902164015-b3e70e2ff481
	github.com/karrick/godirwalk v1.15.6
	github.com/lib/pq v1.8.0 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/yuin/goldmark v1.2.1
	go.uber.org/atomic v1.5.0
	go.uber.org/zap v1.14.0
	golang.org/x/net v0.0.0-20201026091529-146b70c837a4
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	google.golang.org/api v0.29.0
)

replace github.com/jschaf/bibtex => ../bibtex
