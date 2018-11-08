package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/namreg/godown/internal/api"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/grpc"
)

const connectTimeout = 100 * time.Millisecond

//go:generate minimock -i github.com/namreg/godown/client.executor -o ./
type executor interface {
	ExecuteCommand(context.Context, *api.ExecuteCommandRequest, ...grpc.CallOption) (*api.ExecuteCommandResponse, error)
}

//Client is a client that communicates with a server.
type Client struct {
	addrs    []string
	conn     *grpc.ClientConn
	executor executor
}

//New creates a new client with the given servet addresses.
func New(addr string, addrs ...string) (*Client, error) {
	c := &Client{addrs: append([]string{addr}, addrs...)}
	if err := c.tryConnect(); err != nil {
		return nil, fmt.Errorf("could not connect to server: %v", err)
	}
	return c, nil
}

//Close closes the client.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) tryConnect() error {
	var (
		result *multierror.Error
		err    error
		conn   *grpc.ClientConn
	)

	for addrs := c.addrs; len(addrs) > 0; addrs = addrs[1:] {
		ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
		defer cancel()

		if conn, err = grpc.DialContext(ctx, addrs[0], grpc.WithInsecure(), grpc.WithBlock()); err == nil {
			c.conn = conn
			c.executor = api.NewGodownClient(c.conn)
			return nil
		}
		result = multierror.Append(result, err)
	}
	return result.ErrorOrNil()
}

func (c *Client) newExecuteRequest(cmd string, args ...string) *api.ExecuteCommandRequest {
	args = append([]string{cmd}, args...)
	return &api.ExecuteCommandRequest{
		Command: strings.Join(args, " "),
	}
}

//Get gets a value at the given key.
func (c *Client) Get(key string) ScalarResult {
	return c.get(context.Background(), key)
}

//GetWithContext similar to Get but with the context.
func (c *Client) GetWithContext(ctx context.Context, key string) ScalarResult {
	return c.get(ctx, key)
}

func (c *Client) get(ctx context.Context, key string) ScalarResult {
	req := c.newExecuteRequest("GET", key)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}

//Set sets a new value at the given key.
func (c *Client) Set(key, value string) StatusResult {
	return c.set(context.Background(), key, value)
}

//SetWithContext similar to Set but with the context.
func (c *Client) SetWithContext(ctx context.Context, key, value string) StatusResult {
	return c.set(ctx, key, value)
}

func (c *Client) set(ctx context.Context, key, value string) StatusResult {
	req := c.newExecuteRequest("SET", key, value)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return StatusResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newStatusResult(resp)
}

//Del deletes the given key.
func (c *Client) Del(key string) StatusResult {
	return c.del(context.Background(), key)
}

//DelWithContext similar to Del but with context.
func (c *Client) DelWithContext(ctx context.Context, key string) StatusResult {
	return c.del(ctx, key)
}

func (c *Client) del(ctx context.Context, key string) StatusResult {
	req := c.newExecuteRequest("DEL", key)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return StatusResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newStatusResult(resp)
}

//Expire sets expiration of the given key as `now + secs`.
func (c *Client) Expire(key string, secs int) StatusResult {
	return c.expire(context.Background(), key, secs)
}

//ExpireWithContext similar to Expire but with context.
func (c *Client) ExpireWithContext(ctx context.Context, key string, secs int) StatusResult {
	return c.expire(ctx, key, secs)
}

func (c *Client) expire(ctx context.Context, key string, secs int) StatusResult {
	req := c.newExecuteRequest("EXPIRE", key, strconv.Itoa(secs))
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return StatusResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newStatusResult(resp)
}

//GetBit returns the bit value at the offset in the string stored at key.
func (c *Client) GetBit(key string, offset uint64) ScalarResult {
	return c.getBit(context.Background(), key, offset)
}

//GetBitWithContext similar to GetBit but with context.
func (c *Client) GetBitWithContext(ctx context.Context, key string, offset uint64) ScalarResult {
	return c.getBit(ctx, key, offset)
}

func (c *Client) getBit(ctx context.Context, key string, offset uint64) ScalarResult {
	req := c.newExecuteRequest("GETBIT", key, strconv.FormatUint(offset, 10))
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}

//HGet returns the value associated with field in the hash stored at key.
func (c *Client) HGet(key, field string) ScalarResult {
	return c.hget(context.Background(), key, field)
}

//HGetWithContext similar to HGet but with context.
func (c *Client) HGetWithContext(ctx context.Context, key, field string) ScalarResult {
	return c.hget(ctx, key, field)
}

func (c *Client) hget(ctx context.Context, key, field string) ScalarResult {
	req := c.newExecuteRequest("HGET", key, field)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}

//HKeys returns all keys of the map stored at the given key.
func (c *Client) HKeys(key string) ListResult {
	return c.hkeys(context.Background(), key)
}

//HKeysWithContext similar to Hkeys by with context.
func (c *Client) HKeysWithContext(ctx context.Context, key string) ListResult {
	return c.hkeys(ctx, key)
}

func (c *Client) hkeys(ctx context.Context, key string) ListResult {
	req := c.newExecuteRequest("HKEYS", key)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ListResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newListResult(resp)
}

//HSet sets a field in hash at the given key.
func (c *Client) HSet(key, field, value string) StatusResult {
	return c.hset(context.Background(), key, field, value)
}

//HSetWithContext similar to HSet but with context.
func (c *Client) HSetWithContext(ctx context.Context, key, field, value string) StatusResult {
	return c.hset(ctx, key, field, value)
}

func (c *Client) hset(ctx context.Context, key, field, value string) StatusResult {
	req := c.newExecuteRequest("HSET", key, field, value)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return StatusResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newStatusResult(resp)
}

//HVals returns values of a hash stored at the given key.
func (c *Client) HVals(key string) ListResult {
	return c.hvals(context.Background(), key)
}

//HValsWithContext similar to HVals but with context.
func (c *Client) HValsWithContext(ctx context.Context, key string) ListResult {
	return c.hvals(ctx, key)
}

func (c *Client) hvals(ctx context.Context, key string) ListResult {
	req := c.newExecuteRequest("HVALS", key)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ListResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newListResult(resp)
}

//HDel deletes given fields from map stored at the given key.
func (c *Client) HDel(key string, field string, fields ...string) ScalarResult {
	return c.hdel(context.Background(), key, field, fields...)
}

//HDelWithContext similar to HDel but with context.
func (c *Client) HDelWithContext(ctx context.Context, key, field string, fields ...string) ScalarResult {
	return c.hdel(ctx, key, field, fields...)
}

func (c *Client) hdel(ctx context.Context, key, field string, fields ...string) ScalarResult {
	args := append([]string{key, field}, fields...)
	req := c.newExecuteRequest("HDEL", args...)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}

//Keys returns all keys that matched to the given pattern.
func (c *Client) Keys(pattern string) ListResult {
	return c.keys(context.Background(), pattern)
}

//KeysWithContext similar to Keys but with context.
func (c *Client) KeysWithContext(ctx context.Context, pattern string) ListResult {
	return c.keys(ctx, pattern)
}

func (c *Client) keys(ctx context.Context, pattern string) ListResult {
	req := c.newExecuteRequest("KEYS", pattern)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ListResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newListResult(resp)
}

//LIndex returns a value at the index in the list stored at the given key.
func (c *Client) LIndex(key string, index int) ScalarResult {
	return c.lindex(context.Background(), key, index)
}

//LIndexWithContext similar to LIndex but with context.
func (c *Client) LIndexWithContext(ctx context.Context, key string, index int) ScalarResult {
	return c.lindex(ctx, key, index)
}

func (c *Client) lindex(ctx context.Context, key string, index int) ScalarResult {
	req := c.newExecuteRequest("LINDEX", key, strconv.Itoa(index))
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}

//LLen returns a number of the elements in the list stored at the given key.
func (c *Client) LLen(key string) ScalarResult {
	return c.llen(context.Background(), key)
}

//LLenWithContext similar to LLen but with context.
func (c *Client) LLenWithContext(ctx context.Context, key string) ScalarResult {
	return c.llen(ctx, key)
}

func (c *Client) llen(ctx context.Context, key string) ScalarResult {
	req := c.newExecuteRequest("LLEN", key)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}

//LPop removes and returns the first element of the list stored at the given key.
func (c *Client) LPop(key string) ScalarResult {
	return c.lpop(context.Background(), key)
}

//LPopWithContext similar to LPop but with context.
func (c *Client) LPopWithContext(ctx context.Context, key string) ScalarResult {
	return c.lpop(ctx, key)
}

func (c *Client) lpop(ctx context.Context, key string) ScalarResult {
	req := c.newExecuteRequest("LPOP", key)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}

//LPush prepends a new value to the list stored at the given key.
func (c *Client) LPush(key, value string) StatusResult {
	return c.lpush(context.Background(), key, value)
}

//LPushWithContext similar to LPush but with context.
func (c *Client) LPushWithContext(ctx context.Context, key, value string) StatusResult {
	return c.lpush(ctx, key, value)
}

func (c *Client) lpush(ctx context.Context, key, value string) StatusResult {
	req := c.newExecuteRequest("LPUSH", key, value)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return StatusResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newStatusResult(resp)
}

// RPush appends a new value(s) to the list stored at the given key.
func (c *Client) RPush(key, value string, values ...string) StatusResult {
	return c.rpush(context.Background(), key, value, values...)
}

// RPushWithContext similar to RPush but with context.
func (c *Client) RPushWithContext(ctx context.Context, key, value string, values ...string) StatusResult {
	return c.rpush(ctx, key, value, values...)
}

func (c *Client) rpush(ctx context.Context, key, value string, values ...string) StatusResult {
	args := append([]string{key, value}, values...)
	req := c.newExecuteRequest("RPUSH", args...)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return StatusResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newStatusResult(resp)
}

//LRange returns elements from the list stored at the given key.
//Start and stop are zero-based indexes.
func (c *Client) LRange(key string, start, stop int) ListResult {
	return c.lrange(context.Background(), key, start, stop)
}

//LRangeWithContext similar to LRange but with context.
func (c *Client) LRangeWithContext(ctx context.Context, key string, start, stop int) ListResult {
	return c.lrange(ctx, key, start, stop)
}

func (c *Client) lrange(ctx context.Context, key string, start, stop int) ListResult {
	req := c.newExecuteRequest("LRANGE", key, strconv.Itoa(start), strconv.Itoa(stop))
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ListResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newListResult(resp)
}

//LRem removes a given value from the list stored at the given key.
func (c *Client) LRem(key, value string) StatusResult {
	return c.lrem(context.Background(), key, value)
}

//LRemWithContext similar to LRem but with context.
func (c *Client) LRemWithContext(ctx context.Context, key, value string) StatusResult {
	return c.lrem(ctx, key, value)
}

func (c *Client) lrem(ctx context.Context, key, value string) StatusResult {
	req := c.newExecuteRequest("LREM", key, value)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return StatusResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newStatusResult(resp)
}

//Ping returns PONG if no argument is provided, otherwise returns a copy of the argument as a bulk.
func (c *Client) Ping(args ...string) ScalarResult {
	return c.ping(context.Background(), args...)
}

//PingWithContext similar to Ping but with context.
func (c *Client) PingWithContext(ctx context.Context, args ...string) ScalarResult {
	return c.ping(ctx, args...)
}

func (c *Client) ping(ctx context.Context, args ...string) ScalarResult {
	req := c.newExecuteRequest("PING", args...)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}

//SetBit sets or clears the bit at offset in the bitmap stored at key.
func (c *Client) SetBit(key string, offset, value uint64) StatusResult {
	return c.setbit(context.Background(), key, offset, value)
}

//SetBitWithContext similar to SetBit but with context.
func (c *Client) SetBitWithContext(ctx context.Context, key string, offset, value uint64) StatusResult {
	return c.setbit(ctx, key, offset, value)
}

func (c *Client) setbit(ctx context.Context, key string, offset, value uint64) StatusResult {
	req := c.newExecuteRequest("SETBIT", key, strconv.FormatUint(offset, 10), strconv.FormatUint(value, 10))
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return StatusResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newStatusResult(resp)
}

//Strlen returns a length of the string stored at key.
func (c *Client) Strlen(key string) ScalarResult {
	return c.strlen(context.Background(), key)
}

//StrlenWithContext similar to Strlen but with context.
func (c *Client) StrlenWithContext(ctx context.Context, key string) ScalarResult {
	return c.strlen(ctx, key)
}

func (c *Client) strlen(ctx context.Context, key string) ScalarResult {
	req := c.newExecuteRequest("STRLEN", key)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}

//TTL returns the remaining time to live of a key. -1 returns if key does not have timeout.
func (c *Client) TTL(key string) ScalarResult {
	return c.ttl(context.Background(), key)
}

//TTLWithContext similar to TTL but with context.
func (c *Client) TTLWithContext(ctx context.Context, key string) ScalarResult {
	return c.ttl(ctx, key)
}

func (c *Client) ttl(ctx context.Context, key string) ScalarResult {
	req := c.newExecuteRequest("TTL", key)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}

//Type returns a data type of a value stored at key.
func (c *Client) Type(key string) ScalarResult {
	return c.getType(context.Background(), key)
}

//TypeWithContext similar to Type but with context.
func (c *Client) TypeWithContext(ctx context.Context, key string) ScalarResult {
	return c.getType(ctx, key)
}

func (c *Client) getType(ctx context.Context, key string) ScalarResult {
	req := c.newExecuteRequest("TYPE", key)
	resp, err := c.executor.ExecuteCommand(ctx, req)
	if err != nil {
		return ScalarResult{err: fmt.Errorf("could not execute command: %v", err)}
	}
	return newScalarResult(resp)
}
