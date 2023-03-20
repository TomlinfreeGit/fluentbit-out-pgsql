package main

import "C"
import (
	"encoding/json"
	"fmt"
	"go-fb-pgsql/entity"
	"go-fb-pgsql/utils"
	"math/rand"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {}

var rnd = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

var contexts = make(map[string]*Ctx)

type Ctx struct {
	url  string
	conn *gorm.DB
}

type record map[string]interface{}

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	// Gets called only once when the plugin.so is loaded
	return output.FLBPluginRegister(def, "out_pgsql", "golang pgsql output plugin!")
}

//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	// Gets called only once for each instance you have configured.

	ctx := Ctx{}

	pgHost := output.FLBPluginConfigKey(plugin, "pghost")
	if pgHost == "" {
		pgHost = "127.0.0.1"
	}
	pgPort := output.FLBPluginConfigKey(plugin, "pgport")
	if pgPort == "" {
		pgPort = "5432"
	}
	pgUser := output.FLBPluginConfigKey(plugin, "user")
	if pgUser == "" {
		pgUser = "postgres"
	}
	pgPwd := output.FLBPluginConfigKey(plugin, "password")
	pgDb := output.FLBPluginConfigKey(plugin, "database")
	pgTable := output.FLBPluginConfigKey(plugin, "table")
	if pgTable == "" {
		entity.SetTableName("records")
	} else {
		entity.SetTableName(pgTable)
	}

	ctx.url = utils.GetGormPostgresUrl(pgUser, pgPwd, pgHost, pgPort, pgDb)
	conn, err := gorm.Open(postgres.Open(ctx.url), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().Local()
		}})
	if err != nil {
		fmt.Println("connect fail")
		fmt.Println(err.Error())
		id := strconv.Itoa(int(rnd.Int63n(1000)))
		contexts[id] = &ctx
		output.FLBPluginSetContext(plugin, id)
		return output.FLB_OK
	}
	conn.AutoMigrate(&entity.Record{})
	ctx.conn = conn
	id := strconv.Itoa(int(rnd.Int63n(1000)))
	contexts[id] = &ctx

	output.FLBPluginSetContext(plugin, id)

	return output.FLB_OK
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(plugin, data unsafe.Pointer, length C.int, tag *C.char) int {
	// Gets called with a batch of records to be written to an instance.

	id := output.FLBPluginGetContext(plugin).(string)
	ctx := contexts[id]
	dec := output.NewDecoder(data, int(length))

	for {
		ret, ts, rec := output.GetRecord(dec)
		if ret != 0 {
			break
		}

		m, err := decodeRecord(rec)
		if err != nil {
			fmt.Println(err)
			continue
		}
		wg := &sync.WaitGroup{}

		wg.Add(1)
		mstr, _ := json.Marshal(m)
		tagstr := C.GoString(tag)
		var timestamp time.Time
		switch t := ts.(type) {
		case output.FLBTime:
			timestamp = ts.(output.FLBTime).Time
		case uint64:
			timestamp = time.Unix(int64(t), 0)
		default:
			fmt.Println("time provided invalid, defaulting to now.")
			timestamp = time.Now()
		}
		tsstr := timestamp.String()
		fmt.Printf("tag is %s, ts is %s, record is %s", tagstr, tsstr, string(mstr))
		fmt.Println()
		go flushDataToPgsql(mstr, tagstr, timestamp, ctx, wg)

		wg.Wait()
	}

	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func decodeRecord(record map[interface{}]interface{}) (record, error) {
	var err error
	m := make(map[string]interface{})

	for k, v := range record {
		kk, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("bad key type: %T", k)
		}

		switch x := v.(type) {
		case []uint8:
			m[kk] = string(x)
		case map[interface{}]interface{}:
			m[kk], err = decodeRecord(x)
			if err != nil {
				return nil, err
			}
		default:
			m[kk] = x
		}
	}

	return m, nil
}

func flushDataToPgsql(record []byte, tag string, ts time.Time, ctx *Ctx, wg *sync.WaitGroup) {
	defer wg.Done()
	if ctx.conn != nil {
		r := entity.Record{Timestamp: ts, Tag: tag, Data: record}
		ctx.conn.Create(&r)
	}
}
