package patch

import "gitlab.com/marsskom/burro/internal/model"

func ApplyCtxPatch(ctx *model.RequestContext, p *model.CtxPatch) error {
	if p == nil {
		return nil
	}

	if p.IsFinished != nil && *p.IsFinished {
		ctx.Finish()
	}

	return nil
}
