package tikvclient

import (
	. "database/sql/driver"
	"errors"
	"strconv"
	"strings"

	"github.com/journeymidnight/yig/helper"
	. "github.com/journeymidnight/yig/meta/types"
)

func genGcKey(backend BackendType, fsid, poolName, objectId string) []byte {
	backendStr := strconv.Itoa(int(backend))
	return GenKey(TableGcPrefix, poolName, fsid, backendStr, objectId)
}

//gc
func (c *TiKVClient) PutObjectToGarbageCollection(object *Object, tx Tx) error {
	key := genGcKey(object.Backend, object.Location, object.Pool, object.ObjectId)
	gc := GetGcInfoFromObject(object)
	if tx == nil {
		return c.TxPut(key, gc)
	}
	txn := tx.(*TikvTx).tx
	v, err := helper.MsgPackMarshal(gc)
	if err != nil {
		return err
	}
	return txn.Set(key, v)
}

//Key: g\{PoolName}\{Fsid}\{Backend}\{ObjectId}
func (c *TiKVClient) ScanGarbageCollection(limit int) (gcs []GarbageCollection, err error) {
	startKey := GenKey(TableGcPrefix, TableMinKeySuffix)
	endKey := GenKey(TableGcPrefix, TableMaxKeySuffix)
	kvs, err := c.TxScan(startKey, endKey, limit)
	if err != nil {
		return nil, err
	}

	for _, kv := range kvs {
		var gc GarbageCollection
		key, val := kv.K, kv.V
		err = helper.MsgPackUnMarshal(val, &gc)
		sp := strings.Split(string(key), TableSeparator)
		if len(sp) != 5 {
			return nil, errors.New("Invalid gc key:" + string(key))
		}
		gcs = append(gcs, gc)
	}

	return
}

func (c *TiKVClient) RemoveGarbageCollection(garbage GarbageCollection) error {
	key := genGcKey(garbage.Backend, garbage.Location, garbage.Pool, garbage.ObjectId)
	return c.TxDelete(key)
}