package ctx

import (
	"context"

	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	dbKey ctxKey = iota
	logKey
	userIDKey
	secretKey
)

func DBProvider(q data.UsersQ) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, dbKey, q)
	}
}

func DB(ctx context.Context) data.UsersQ {
	return ctx.Value(dbKey).(data.UsersQ).New()
}

func LoggerProvider(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logKey, entry)
	}
}

func Logger(ctx context.Context) *logan.Entry {
	return ctx.Value(logKey).(*logan.Entry)
}

func UserIDProvider(id int) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, userIDKey, id)
	}
}

func UserID(ctx context.Context) int {
	return ctx.Value(userIDKey).(int)
}

func JWTSecretProvider(secret []byte) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, secretKey, secret)
	}
}

func JWTSecret(ctx context.Context) []byte {
	return ctx.Value(secretKey).([]byte)
}
