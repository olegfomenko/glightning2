package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/olegfomenko/glightning2/clnrpc"
)

// ErrDatastoreNotFound indicates that an exact datastore key does not exist.
var ErrDatastoreNotFound = errors.New("datastore entry not found")

// DatastoreClient stores and retrieves JSON values through Core Lightning's
// datastore RPC methods.
type DatastoreClient[T any] struct {
	client *clnrpc.Client
}

// GetDatastore returns a new datastore helper backed by the plugin's CLN client.
func GetDatastore[T any](client *clnrpc.Client) *DatastoreClient[T] {
	return &DatastoreClient[T]{client: client}
}

// Save serializes value as JSON and stores it under key using mode.
// Generation is optional and enables atomic updates when supplied.
func (d *DatastoreClient[T]) Save(ctx context.Context, key []string, value T, mode clnrpc.DatastoreMode, generation ...uint64) error {
	_, err := d.SaveRaw(ctx, key, value, mode, generation...)
	return err
}

// SaveRaw serializes value as JSON and stores it under key using mode.
// Generation is optional and enables atomic updates when supplied.
// Also returns the raw response from datastore call
func (d *DatastoreClient[T]) SaveRaw(ctx context.Context, key []string, value T, mode clnrpc.DatastoreMode, generation ...uint64) (*clnrpc.DatastoreResponse, error) {
	if len(key) == 0 {
		return nil, errors.New("datastore key is required")
	}
	if mode == "" {
		return nil, errors.New("datastore mode is required")
	}
	if len(generation) > 1 {
		return nil, errors.New("datastore accepts at most one generation")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	request := clnrpc.DatastoreRequest{
		Key:    key,
		Mode:   mode,
		String: string(data),
	}
	if len(generation) == 1 {
		request.Generation = generation[0]
	}
	return d.client.Datastore(ctx, request)
}

// Get loads and decodes the value at the exact key.
func (d *DatastoreClient[T]) Get(ctx context.Context, key []string) (T, error) {
	value, _, err := d.GetRaw(ctx, key)
	return value, err
}

// GetRaw loads and decodes the value at the exact key.
// Also returns the raw response from listdatastore call
func (d *DatastoreClient[T]) GetRaw(ctx context.Context, key []string) (T, *clnrpc.ListDatastoreResponse, error) {
	var value T
	if len(key) == 0 {
		return value, nil, errors.New("datastore key is required")
	}

	values, response, err := d.ListRaw(ctx, key)
	if err != nil {
		return value, nil, err
	}

	if len(values) > 1 {
		return value, nil, errors.New("found more then one entry")
	}

	if len(values) == 0 {
		return value, nil, ErrDatastoreNotFound
	}

	return values[0], response, ErrDatastoreNotFound
}

// List loads and decodes all datastore entries below the optional key prefix.
func (d *DatastoreClient[T]) List(ctx context.Context, key []string) ([]T, error) {
	values, _, err := d.ListRaw(ctx, key)
	return values, err
}

// ListRaw loads and decodes all datastore entries below the optional key prefix.
// Also returns the raw response from listdatastore call
func (d *DatastoreClient[T]) ListRaw(ctx context.Context, key []string) ([]T, *clnrpc.ListDatastoreResponse, error) {
	response, err := d.client.ListDatastore(ctx, clnrpc.ListDatastoreRequest{Key: key})
	if err != nil {
		return nil, nil, err
	}

	values := make([]T, len(response.Datastore))
	for i := range response.Datastore {
		if err := json.Unmarshal([]byte(response.Datastore[i].String), &values[i]); err != nil {
			return nil, nil, err
		}
	}
	return values, response, nil
}

// Delete removes the value at key. Generation is optional and enables an atomic delete when supplied.
func (d *DatastoreClient[T]) Delete(ctx context.Context, key []string, generation ...uint64) error {
	_, err := d.DeleteRaw(ctx, key, generation...)
	return err
}

// DeleteRaw removes the value at key. Generation is optional and enables an atomic delete when supplied.
// Also returns the raw response from deldatastore call
func (d *DatastoreClient[T]) DeleteRaw(ctx context.Context, key []string, generation ...uint64) (*clnrpc.DelDatastoreResponse, error) {
	if len(key) == 0 {
		return nil, errors.New("datastore key is required")
	}
	if len(generation) > 1 {
		return nil, errors.New("datastore accepts at most one generation")
	}

	request := clnrpc.DelDatastoreRequest{Key: key}
	if len(generation) == 1 {
		request.Generation = generation[0]
	}
	return d.client.DelDatastore(ctx, request)
}

// Exists reports whether an exact datastore key exists.
func (d *DatastoreClient[T]) Exists(ctx context.Context, key []string) (bool, error) {
	_, _, err := d.GetRaw(ctx, key)
	if errors.Is(err, ErrDatastoreNotFound) {
		return false, nil
	}
	return err == nil, err
}
