package plugin

import "gitlab.com/marsskom/burro/internal/pluginapi"

type runtime struct {
	artifacts pluginapi.ArtifactStore
	data      pluginapi.DataStore
	kv        pluginapi.KeyValueStore
	eventBus  pluginapi.EventBus
	log       pluginapi.Logger
}

func NewRuntime(
	a pluginapi.ArtifactStore,
	d pluginapi.DataStore,
	kv pluginapi.KeyValueStore,
	e pluginapi.EventBus,
	l pluginapi.Logger,
) pluginapi.Runtime {
	return &runtime{
		artifacts: a,
		data:      d,
		kv:        kv,
		eventBus:  e,
		log:       l,
	}
}

func (r *runtime) Artifacts() pluginapi.ArtifactStore { return r.artifacts }
func (r *runtime) Data() pluginapi.DataStore          { return r.data }
func (r *runtime) KV() pluginapi.KeyValueStore        { return r.kv }
func (r *runtime) Events() pluginapi.EventBus         { return r.eventBus }
func (r *runtime) Log() pluginapi.Logger              { return r.log }
