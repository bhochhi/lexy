package orchestrator

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bhochhi/lexy/codehook"
	"github.com/bhochhi/lexy/nlu"
	"github.com/bhochhi/lexy/session"
)

// Contract types based on instructions/workflow template

type Workflow struct {
	Intent      string `json:"intent"`
	NLUIntent   string `json:"nluIntent"`
	Description string `json:"description"`
	Steps       []Step `json:"steps"`
}

type Step struct {
	StepID     string       `json:"stepId"`
	Type       string       `json:"type"`    // validation | slot | message | fallback
	SkipNLU    bool         `json:"skipNLU"` // default false; if true, do not call NLU on user input for this step
	Prompts    []string     `json:"prompts"` // first one is default, others for retries prompts
	Options    *StepOptions `json:"options"`
	Contents   []Content    `json:"contents"` // for message/fallback steps as final message
	Validation *Validation  `json:"validation"`
	OnSuccess  any          `json:"onSuccess"` // string or map[string]string or special resolver name
	OnFailure  string       `json:"onFailure"`
	MaxRetries int          `json:"maxRetries"`
}

type StepOptions struct {
	Static    []session.Option `json:"static"`
	Functions []FunctionSpec   `json:"functions"`
}

type FunctionSpec struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type Content struct {
	Text string `json:"text"`
}

type Validation struct {
	Pre  *ValidationCall `json:"pre"`
	Post *ValidationCall `json:"post"`
}

type ValidationCall struct {
	Function  string `json:"function"`
	OnSuccess any    `json:"onSuccess"` // string | map[string]string | special resolver
	OnFailure string `json:"onFailure"`
}

// Session contracts (shared with session.go)

type SessionStore interface {
	Get(clientID string) (*session.Session, bool)
	Save(*session.Session)
}

// Code hooks registry contract (intent + domain hooks) now lives in codehook.Hook

// NLU contract

type NLU interface {
	DetectIntent(text, clientID string) (intent string, slots map[string]string, err error)
}

// Public Orchestrator

type Orchestrator struct {
	Store   SessionStore
	NLU     NLU
	Funcs   codehook.Hook
	WorkDir string // path to workflow dir
}

func New(store SessionStore, nluClient nlu.Client, funcs codehook.Hook, workDir string) *Orchestrator {
	return &Orchestrator{Store: store, NLU: nluClient, Funcs: funcs, WorkDir: workDir}
}

// Execute handles a user input and returns response text + options.
func (o *Orchestrator) Execute(clientID, userText string) (string, []session.Option, error) {
	sess, ok := o.Store.Get(clientID)
	if !ok {
		// New session: run NLU (mocked) and initialize workflow
		nluIntent, slots, _ := o.NLU.DetectIntent(userText, clientID)
		wf, err := o.loadWorkflowByNLU(nluIntent)
		if err != nil {
			return "", nil, err
		}
		sess = &session.Session{
			ClientID:      clientID,
			CurrentStepID: firstStepID(wf),
			StepsState:    map[string]*session.StepState{},
			// Store canonical internal intent under "intent"; keep nluIntent for reference
			Context: map[string]any{"intent": wf.Intent, "nluIntent": nluIntent, "clientId": clientID},
		}
		for k, v := range slots {
			sess.Context[k] = v
		}
	}

	wf, err := o.loadWorkflow(fmt.Sprintf("%v", sess.Context["nluIntent"]))
	if err != nil {
		return "", nil, err
	}

	// Main loop: process current step based on type
	remainingInput := userText
	for i := 0; i < 20; i++ { // guard to avoid infinite loops in malformed configs
		step, found := findStep(wf, sess.CurrentStepID)
		if !found {
			return "", nil, fmt.Errorf("step not found: %s", sess.CurrentStepID)
		}

		state := ensureStepState(sess, step)

		switch strings.ToLower(step.Type) {
		case "validation":
			// Pre validation func
			if step.Validation != nil && step.Validation.Pre != nil && !state.ValidationDone {
				result, err := o.Funcs.Invoke(step.Validation.Pre.Function, sess.Context, resolveArgs(sess.Context, map[string]any{}))
				if err != nil {
					// on failure
					next := step.Validation.Pre.OnFailure
					sess.CurrentStepID = next
					sess.Options = nil
					resetState(state)
					continue
				}
				// Handle hasData/hasNoData branching
				if v, ok := step.Validation.Pre.OnSuccess.(map[string]any); ok {
					if getBool(result, "hasData") {
						sess.CurrentStepID = fmt.Sprintf("%v", v["hasData"])
					} else {
						sess.CurrentStepID = fmt.Sprintf("%v", v["hasNoData"])
					}
					sess.Options = nil
					resetState(state)
					state.ValidationDone = true
					continue
				}
			}
			// If no pre validation or done, go to onSuccess or configured next step
			if s, ok := step.OnSuccess.(string); ok && s != "" {
				sess.CurrentStepID = s
				sess.Options = nil
				resetState(state)
				continue
			}
			return "Validation step configured incorrectly", nil, nil

		case "slot":
			// Render prompt based on retry counter
			prompt := ""
			if len(step.Prompts) > 0 {
				idx := state.LastPromptIdx
				if idx >= len(step.Prompts) {
					idx = len(step.Prompts) - 1
				}
				prompt = step.Prompts[idx]
			}

			// collect options (static + dynamic)
			opts, err := o.collectOptions(sess, step)
			if err != nil {
				// any function failure routes to onFailure
				sess.CurrentStepID = step.OnFailure
				sess.Options = nil
				resetState(state)
				continue
			}

			// If remaining input provided, capture it for this step only
			if remainingInput != "" {
				state.CapturedSlot = remainingInput
				remainingInput = ""
			}

			// post validation if defined
			if step.Validation != nil && step.Validation.Post != nil {
				_, err := o.Funcs.Invoke(step.Validation.Post.Function, sess.Context, map[string]any{"value": state.CapturedSlot})
				if err != nil {
					// retry
					state.RetryCount++
					state.LastPromptIdx++
					if state.RetryCount > maxRetries(step, state) {
						sess.CurrentStepID = step.OnFailure
						sess.Options = nil
						resetState(state)
						continue
					}
					// stay on same step, return prompt + options
					sess.Options = opts
					o.Store.Save(sess)
					return prompt, opts, nil
				}

				// success branching: map or string
				if m, ok := step.Validation.Post.OnSuccess.(map[string]any); ok {
					next := fmt.Sprintf("%v", m[state.CapturedSlot])
					if next == "" {
						// stay or fallback
						sess.CurrentStepID = step.OnFailure
					} else {
						sess.CurrentStepID = next
					}
				} else if s, ok := step.Validation.Post.OnSuccess.(string); ok {
					// if s is a step id, move; if it's a resolver function, call to get nextStep
					if hasStepID(wf, s) || strings.EqualFold(s, "end") {
						sess.CurrentStepID = s
					} else {
						// treat as function name to resolve next step
						res, err := o.Funcs.Invoke(s, sess.Context, map[string]any{"value": state.CapturedSlot})
						if err != nil {
							sess.CurrentStepID = step.OnFailure
						} else if ns, ok := res["nextStep"].(string); ok && ns != "" {
							sess.CurrentStepID = ns
						} else {
							sess.CurrentStepID = step.OnFailure
						}
					}
				} else {
					sess.CurrentStepID = step.OnFailure
				}

				// reset after success
				resetState(state)
				sess.Options = nil
				continue
			}

			// No post-validation: just return prompt + options and wait for input
			sess.Options = opts
			o.Store.Save(sess)
			if prompt == "" {
				prompt = "Please provide the requested information."
			}
			return prompt, opts, nil

		case "message", "fallback":
			// Render message and options; then end or move
			text := ""
			if len(step.Contents) > 0 {
				text = renderTemplate(step.Contents[0].Text, sess.Context)
			}
			opts, err := o.collectOptions(sess, step)
			if err != nil {
				// function failure -> onFailure
				sess.CurrentStepID = step.OnFailure
				sess.Options = nil
				continue
			}

			sess.Options = opts
			if s, ok := step.OnSuccess.(string); ok && strings.EqualFold(s, "end") {
				o.Store.Save(sess)
				return text, opts, nil
			}
			if s, ok := step.OnSuccess.(string); ok && s != "" {
				sess.CurrentStepID = s
				continue
			}
			// default end
			o.Store.Save(sess)
			return text, opts, nil
		default:
			return "Unsupported step type", nil, nil
		}
	}

	return "Workflow did not complete", nil, nil
}

func (o *Orchestrator) collectOptions(sess *session.Session, step Step) ([]session.Option, error) {
	var options []session.Option
	if step.Options != nil {
		options = append(options, step.Options.Static...)
		for _, f := range step.Options.Functions {
			args := resolveArgs(sess.Context, f.Args)
			res, err := o.Funcs.Invoke(f.Name, sess.Context, args)
			if err != nil {
				return nil, err
			}
			if list, ok := res["options"].([]session.Option); ok {
				options = append(options, list...)
			} else if list2, ok := res["options"].([]map[string]any); ok {
				for _, it := range list2 {
					options = append(options, session.Option{Text: fmt.Sprintf("%v", it["text"]), URL: fmt.Sprintf("%v", it["url"])})
				}
			}
		}
	}
	return options, nil
}

func (o *Orchestrator) loadWorkflow(intent string) (*Workflow, error) {
	if intent == "" {
		return nil, errors.New("missing intent")
	}
	p := filepath.Join(o.WorkDir, fmt.Sprintf("%s.json", intent))
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var wf Workflow
	if err := json.Unmarshal(b, &wf); err != nil {
		return nil, err
	}
	return &wf, nil
}

// loadWorkflowByNLU locates a workflow using the business/NLU intent name.
// Strategy:
// 1) If a file named <nluIntent>.json exists, load it.
// 2) Else, scan all workflow JSONs and pick the one whose nluIntent matches.
// 3) Else, fall back to using the nluIntent as the internal intent (legacy behavior).
func (o *Orchestrator) loadWorkflowByNLU(nluIntent string) (*Workflow, error) {
	if nluIntent == "" {
		return nil, errors.New("missing nlu intent")
	}
	// Fast path: filename match
	p := filepath.Join(o.WorkDir, fmt.Sprintf("%s.json", nluIntent))
	if b, err := os.ReadFile(p); err == nil {
		var wf Workflow
		if err := json.Unmarshal(b, &wf); err != nil {
			return nil, err
		}
		return &wf, nil
	}
	// Scan all workflows for nluIntent match
	entries, err := os.ReadDir(o.WorkDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(o.WorkDir, e.Name()))
		if err != nil {
			continue
		}
		// Use a lightweight struct to peek fields
		var meta struct {
			Intent    string `json:"intent"`
			NLUIntent string `json:"nluIntent"`
		}
		if err := json.Unmarshal(b, &meta); err != nil {
			continue
		}
		if strings.EqualFold(meta.NLUIntent, nluIntent) {
			var wf Workflow
			if err := json.Unmarshal(b, &wf); err != nil {
				return nil, err
			}
			return &wf, nil
		}
	}
	// Fallback: treat nluIntent as internal intent (legacy)
	return o.loadWorkflow(nluIntent)
}

func firstStepID(wf *Workflow) string {
	if len(wf.Steps) == 0 {
		return ""
	}
	return wf.Steps[0].StepID
}

func findStep(wf *Workflow, id string) (Step, bool) {
	for _, s := range wf.Steps {
		if s.StepID == id {
			return s, true
		}
	}
	return Step{}, false
}

func ensureStepState(sess *session.Session, step Step) *session.StepState {
	st, ok := sess.StepsState[step.StepID]
	if !ok {
		st = &session.StepState{MaxRetries: step.MaxRetries}
		sess.StepsState[step.StepID] = st
	}
	if st.MaxRetries == 0 {
		st.MaxRetries = step.MaxRetries
	}
	return st
}

func resetState(s *session.StepState) {
	s.RetryCount = 0
	s.LastPromptIdx = 0
	s.CapturedSlot = ""
	s.ValidationDone = false
}

func maxRetries(step Step, state *session.StepState) int {
	if state.MaxRetries > 0 {
		return state.MaxRetries
	}
	if step.MaxRetries > 0 {
		return step.MaxRetries
	}
	return 0
}

// simplistic handlebars-style templating for {{context.key}}
func renderTemplate(tpl string, ctx map[string]any) string {
	out := tpl
	for k, v := range ctx {
		needle := fmt.Sprintf("{{context.%s}}", k)
		out = strings.ReplaceAll(out, needle, fmt.Sprintf("%v", v))
	}
	return out
}

func resolveArgs(ctx map[string]any, args map[string]any) map[string]any {
	res := map[string]any{}
	for k, v := range args {
		if s, ok := v.(string); ok {
			res[k] = renderTemplate(s, ctx)
		} else {
			res[k] = v
		}
	}
	return res
}

func getBool(m map[string]any, key string) bool {
	if b, ok := m[key].(bool); ok {
		return b
	}
	return false
}

func hasStepID(wf *Workflow, id string) bool {
	if strings.EqualFold(id, "end") || id == "" {
		return false
	}
	for _, s := range wf.Steps {
		if s.StepID == id {
			return true
		}
	}
	return false
}
