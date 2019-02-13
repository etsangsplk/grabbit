package builder

import (
	"fmt"
	"go/types"
	"sync"

	"github.com/rhinof/grabbit/gbus"
	"github.com/rhinof/grabbit/gbus/saga"

	"github.com/rhinof/grabbit/gbus/saga/stores"
	"github.com/rhinof/grabbit/gbus/serialization"
	"github.com/rhinof/grabbit/gbus/tx"
	"github.com/streadway/amqp"
)

type defaultBuilder struct {
	handlers         []types.Type
	connStr          string
	purgeOnStartup   bool
	sagaStoreConnStr string
	txnl             bool
	txConnStr        string
	txnlProvider     string
	workerNum        uint
	serializer       gbus.MessageEncoding
	dlx              string
}

func (builder *defaultBuilder) Build(svcName string) gbus.Bus {

	gb := &gbus.DefaultBus{
		AmqpConnStr:          builder.connStr,
		SvcName:              svcName,
		PurgeOnStartup:       builder.purgeOnStartup,
		ConnErrors:           make(chan *amqp.Error),
		DelayedSubscriptions: [][]string{},
		HandlersLock:         &sync.Mutex{},
		IsTxnl:               builder.txnl,
		MsgHandlers:          make(map[string][]gbus.MessageHandler),
		Serializer:           builder.serializer,
		DLX:                  builder.dlx}

	if builder.workerNum < 1 {
		gb.WorkerNum = 1
	} else {
		gb.WorkerNum = builder.workerNum
	}
	var sagaStore saga.Store
	if builder.txnl {
		gb.IsTxnl = true
		switch builder.txnlProvider {
		case "pg":
			pgtx, err := tx.NewPgProvider(builder.txConnStr)
			if err != nil {
				panic(err)
			}
			gb.TxProvider = pgtx
			sagaStore = stores.NewPgStore(gb.SvcName, pgtx)

		default:
			error := fmt.Errorf("no provider found for passed in value %v", builder.txnlProvider)
			panic(error)
		}
	} else {
		sagaStore = stores.NewInMemoryStore()
	}

	gb.Glue = saga.NewGlue(gb, sagaStore, svcName)
	return gb
}

func (builder *defaultBuilder) PurgeOnStartUp() gbus.Builder {
	builder.purgeOnStartup = true
	return builder
}

func (builder *defaultBuilder) WithOutbox(connStr string) gbus.Builder {

	//TODO: Add outbox suppoert to builder
	return builder
}

func (builder *defaultBuilder) WithDeadlettering(deadletterExchange string) gbus.Builder {

	builder.dlx = deadletterExchange
	//TODO: Add outbox suppoert to builder
	return builder
}

func (builder *defaultBuilder) WorkerNum(workers uint) gbus.Builder {
	builder.workerNum = workers
	return builder
}

/*
	WithSagas configures the bus to work with Sagas.
	sagaStoreConnStr: the connection string to the saga store

	Supported Saga Stores and the format of the connection string to use:
	PostgreSQL: "PostgreSQL;User ID=root;Password=myPassword;Host=localhost;Port=5432;Database=myDataBase;"
	In Memory:  ""
*/
func (builder *defaultBuilder) WithSagas(sagaStoreConnStr string) gbus.Builder {
	builder.sagaStoreConnStr = sagaStoreConnStr
	return builder
}

func (builder *defaultBuilder) Txnl(provider, connStr string) gbus.Builder {
	builder.txnl = true
	builder.txConnStr = connStr
	builder.txnlProvider = provider
	return builder
}

func (builder *defaultBuilder) WithSerializer(serializer gbus.MessageEncoding) gbus.Builder {

	builder.serializer = serializer
	return builder
}

//New :)
func New() Nu {
	return Nu{}
}

//Nu is the new New
type Nu struct {
}

//Bus inits a new BusBuilder
func (Nu) Bus(brokerConnStr string) gbus.Builder {
	return &defaultBuilder{
		connStr:    brokerConnStr,
		serializer: serialization.NewGobSerializer(),
	}
}
