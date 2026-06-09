package actions

import (
	"fmt"
	"net/http"

	"gitlab.com/marsskom/burro/internal/pluginapi"
	"gitlab.com/marsskom/burro/plugins/policy/response"
)

func Execute(log pluginapi.Logger, rules []ActionRule, req *http.Request) *http.Response {
	if len(rules) == 0 {
		return nil
	}

	for _, r := range rules {
		if !ActionMatch(r.Match, req) {
			continue
		}

		for _, a := range r.Action {
			resp, stop, err := ApplyAction(log, a, req)
			if err != nil {
				return response.InternalError(err)
			}

			if resp != nil {
				return resp
			}

			if stop {
				return nil
			}
		}

		if r.OnMatch == OnMatchStop {
			break
		}
	}

	return nil
}

func ApplyAction(log pluginapi.Logger, a Action, req *http.Request) (*http.Response, bool, error) {
	log.Debug("applying operation", "operation", a.Operation)

	switch a.Operation {
	case OpDeny:
		log.Debug("request forbidden")

		return response.Forbidden(), true, nil

	case OpAllow:
		log.Debug("request allowed")

		return nil, true, nil

	case OpSetHeader:
		args, err := decode[SetHeaderArgs](a.Args)
		if err != nil {
			return nil, true, err
		}

		log.Debug("set header", "name", args.Name, "value", args.Value)

		req.Header.Set(args.Name, args.Value)

		return nil, false, nil

	case OpRemoveHeader:
		args, err := decode[RemoveHeaderArgs](a.Args)
		if err != nil {
			return nil, true, err
		}

		log.Debug("remove headers", "names", args.Names)

		for _, n := range args.Names {
			req.Header.Del(n)
		}

		return nil, false, nil

	case OpRedactBody:
		args, err := decode[RedactBodyArgs](a.Args)
		if err != nil {
			return nil, true, err
		}

		log.Debug("remove body fields", "fields", args.Fields)

		body, err := readRequestBody(req)
		if err != nil {
			return nil, true, err
		}

		for _, f := range args.Fields {
			body = redactJSONField(body, f)
		}

		replaceRequestBody(req, body)

		return nil, false, nil

	case OpLog:
		args, err := decode[LogArgs](a.Args)
		if err != nil {
			return nil, true, err
		}

		log.Debug("log operation", "level", args.Level, "message", args.Message)

		sendToLog(log, args.Level, args.Message)

		return nil, false, nil
	}

	return nil, true, fmt.Errorf("unknown operation: '%s'", a.Operation)
}

func sendToLog(log pluginapi.Logger, level, msg string) {
	switch level {
	case "trace":
		log.Trace(msg)
	case "debug":
		log.Debug(msg)
	case "info":
		log.Info(msg)
	case "warn":
		log.Warn(msg)
	case "error":
		log.Error(msg)
	case "audit":
		log.Audit(msg)
	default:
		log.Info(msg)
	}
}
