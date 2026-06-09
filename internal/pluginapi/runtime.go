package pluginapi

type Runtime interface {
	Artifacts() ArtifactStore
	Data() DataStore
	KV() KeyValueStore
	Events() EventBus
	Log() Logger
}
