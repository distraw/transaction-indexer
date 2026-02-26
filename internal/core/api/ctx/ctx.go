package ctx

import (
	"context"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/distraw/transaction-indexer/internal/core/indexer"
	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	storageKey ctxKey = iota
	logKey
	userIDKey
	secretKey
	indexerKey
	rpcKey
)

func StorageProvider(s data.Storage) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, storageKey, s)
	}
}

func Storage(ctx context.Context) data.Storage {
	return ctx.Value(storageKey).(data.Storage).New()
}

func LoggerProvider(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logKey, entry)
	}
}

func Logger(ctx context.Context) *logan.Entry {
	return ctx.Value(logKey).(*logan.Entry)
}

func UserIDProvider(id *int) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, userIDKey, id)
	}
}

func UserID(ctx context.Context) *int {
	return ctx.Value(userIDKey).(*int)
}

func JWTSecretProvider(secret []byte) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, secretKey, secret)
	}
}

func JWTSecret(ctx context.Context) []byte {
	return ctx.Value(secretKey).([]byte)
}

func IndexerProvider(indexer indexer.Indexer) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, indexerKey, indexer)
	}
}

func Indexer(ctx context.Context) indexer.Indexer {
	return ctx.Value(indexerKey).(indexer.Indexer)
}

func RPCProvider(rpc *rpcclient.Client) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, rpcKey, rpc)
	}
}

func RPC(ctx context.Context) *rpcclient.Client {
	return ctx.Value(rpcKey).(*rpcclient.Client)
}
