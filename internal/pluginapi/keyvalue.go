package pluginapi

type KeyValueStore interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	List(prefix string) (map[string][]byte, error)
}
