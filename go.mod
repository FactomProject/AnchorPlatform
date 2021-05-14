module github.com/FactomProject/AnchorPlatform

go 1.15

require (
	github.com/AccumulateNetwork/SMT v0.0.12
	github.com/AdamSLevy/jsonrpc2/v14 v14.1.0
	github.com/FactomProject/btcutil v0.0.0-20200312214114-5fd3eaf71bd2 // indirect
	github.com/FactomProject/ed25519 v0.0.0-20150814230546-38002c4fe7b6
	github.com/FactomProject/factom v0.4.0
	github.com/FactomProject/factomd v1.13.0 // indirect
	github.com/FactomProject/serveridentity v0.0.0-20180611231115-cf42d2aa8deb
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/dustin/go-humanize v1.0.0
	github.com/jinzhu/configor v1.2.1
	github.com/mcuadros/go-defaults v1.2.0
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf // indirect
	golang.org/x/net v0.0.0-20210510120150-4163338589ed // indirect
	golang.org/x/sys v0.0.0-20210511113859-b0526f3d8744 // indirect
)

replace github.com/AdamSLevy/jsonrpc2 => github.com/AdamSLevy/jsonrpc2 v1.1.2-0.20210408185727-2d689058f1f8
