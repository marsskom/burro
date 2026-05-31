package pluginapi

type Runtime interface {
	Artifacts() ArtifactStore
	Data() ArtifactStore
	KV() KeyValueStore
	Events() EventBus
	Log() Logger
}
