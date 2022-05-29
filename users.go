package main

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func Lookup(ctx context.Context, username string) (string, error) {
	value, err := rdb.Get(ctx, username).Result()

	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return value, nil
}

func Iterate(ctx context.Context, match string, withValues bool, maxRows int) ([]*KeyValuePair, error) {
	var err error
	log, ok := ctx.Value(loggerValue).(*zap.Logger)
	if !ok {
		log = zap.L()
	}

	result := make([]*KeyValuePair, 0)
	var found int
	var cursor uint64
	for {
		var scanResult []string
		scanResult, cursor, err = rdb.Scan(ctx, cursor, match, int64(maxRows)).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range scanResult {
			result = append(result, &KeyValuePair{Key: key})
		}
		found += len(scanResult)
		if cursor == 0 || (maxRows > 0 && found >= maxRows) {
			break
		}
	}

	if withValues {
		keys := make([]string, len(result))
		for i := 0; i < len(result); i++ {
			keys[i] = result[i].Key
		}
		var values []interface{}
		values, err = rdb.MGet(ctx, keys...).Result()
		for i := 0; i < len(result); i++ {
			key := result[i].Key
			valueI := values[i]
			if valueI == nil {
				log.Warn("Key has disappeared between SCAN and MGET", zap.String("Key", key))
				continue
			}
			value, ok := values[i].(string)
			if !ok {
				log.Warn("Key contains a non-string value", zap.String("Key", key))
				continue
			}
			result[i].Value = value
		}
	}

	return result, nil
}
