package frontend

import "cloudNativeGo/core"

type FrontEnd interface {
	Start(kv *core.KeyValueStore) error
}
