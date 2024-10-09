module cosmossdk.io/indexer/postgres/testing

go 1.23

require (
	cosmossdk.io/indexer/postgres v0.0.0-00010101000000-000000000000
	cosmossdk.io/schema v0.3.0
	cosmossdk.io/schema/testing v0.0.0
	github.com/fergusstrange/embedded-postgres v1.29.0
	github.com/hashicorp/consul/sdk v0.16.1
	github.com/jackc/pgx/v5 v5.7.1
	github.com/stretchr/testify v1.9.0
	gotest.tools/v3 v3.5.1
)

require (
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	pgregory.net/rapid v1.1.0 // indirect
)

replace cosmossdk.io/indexer/postgres => ../.

replace cosmossdk.io/schema => ../../../schema

replace cosmossdk.io/schema/testing => ../../../schema/testing
