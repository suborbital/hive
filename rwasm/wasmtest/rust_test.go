package wasmtest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/suborbital/reactr/request"
	"github.com/suborbital/reactr/rt"
	"github.com/suborbital/reactr/rwasm"
)

func TestWasmRunnerWithFetch(t *testing.T) {
	r := rt.New()

	// test a WASM module that is loaded directly instead of through the bundle
	doWasm := r.Register("wasm", rwasm.NewRunner("../testdata/fetch/fetch.wasm"))

	res, err := doWasm("https://1password.com").Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	if len(res.([]byte)) < 100 {
		t.Errorf("expected 1password.com HTML, got %q", string(res.([]byte)))
	}

	if string(res.([]byte))[:100] != "{\"args\":{},\"data\":{\"message\":\"testing the echo!\"},\"files\":{},\"form\":{},\"headers\":{\"x-forwarded-proto" {
		t.Errorf("expected echo response, got %q", string(res.([]byte))[:100])
	}
}

func TestGraphQLRunner(t *testing.T) {
	r := rt.New()

	// test a WASM module that is loaded directly instead of through the bundle
	doWasm := r.Register("wasm", rwasm.NewRunner("../testdata/rs-graqhql/rs-graqhql.wasm"))

	_, err := doWasm("").Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}
}

func TestWasmRunnerReturnError(t *testing.T) {
	job := rt.NewJob("return-err", "")

	_, err := sharedRT.Do(job).Then()
	if err == nil {
		t.Error("expected error, got none")
		return
	}

	runErr := &rt.RunErr{}
	if !errors.As(err, runErr) || runErr.Error() != `{"code":400,"message":"job failed"}` {
		t.Error("expected RunErr JSON, got", err.Error())
	}
}

func TestWasmRunnerWithRequest(t *testing.T) {
	r := rt.New()

	// using a Rust module
	doWasm := r.Register("wasm", rwasm.NewRunner("../testdata/log/log.wasm"))

	body := testBody{
		Username: "cohix",
	}

	bodyJSON, _ := json.Marshal(body)

	req := &request.CoordinatedRequest{
		Method: "GET",
		URL:    "/hello/world",
		ID:     uuid.New().String(),
		Body:   bodyJSON,
		State: map[string][]byte{
			"hello": []byte("what is up"),
		},
	}

	reqJSON, err := req.ToJSON()
	if err != nil {
		t.Error("failed to ToJSON", err)
	}

	res, err := doWasm(reqJSON).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	resp := &request.CoordinatedResponse{}
	if err := json.Unmarshal(res.([]byte), resp); err != nil {
		t.Error("failed to Unmarshal response")
	}

	if string(resp.Output) != "hello what is up" {
		t.Error(fmt.Errorf("expected 'hello, what is up', got %s", string(res.([]byte))))
	}
}

func TestContentType(t *testing.T) {
	body := testBody{
		Username: "cohix",
	}

	bodyJSON, _ := json.Marshal(body)

	req := &request.CoordinatedRequest{
		Method: "GET",
		URL:    "/hello/world",
		ID:     uuid.New().String(),
		Body:   bodyJSON,
		State: map[string][]byte{
			"hello": []byte("what is up"),
		},
	}

	reqJSON, err := req.ToJSON()
	if err != nil {
		t.Error("failed to ToJSON", err)
	}

	job := rt.NewJob("content-type", reqJSON)

	res, err := sharedRT.Do(job).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	resp := &request.CoordinatedResponse{}
	if err := json.Unmarshal(res.([]byte), resp); err != nil {
		t.Error("failed to Unmarshal response")
	}

	if resp.RespHeaders["Content-Type"] != "application/json" {
		t.Error("unexpected ctype, actually is", resp.RespHeaders["Content-Type"])
	}
}

func TestWasmRunnerDataConversion(t *testing.T) {
	r := rt.New()

	doWasm := r.Register("wasm", rwasm.NewRunner("../testdata/hello-echo/hello-echo.wasm"))

	res, err := doWasm("my name is joe").Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
	}

	if string(res.([]byte)) != "hello my name is joe" {
		t.Error(fmt.Errorf("expected 'hello my name is joe', got %s", string(res.([]byte))))
	}
}

func TestWasmRunnerGroup(t *testing.T) {
	r := rt.New()

	doWasm := r.Register("wasm", rwasm.NewRunner("../testdata/hello-echo/hello-echo.wasm"))

	grp := rt.NewGroup()
	for i := 0; i < 50000; i++ {
		grp.Add(doWasm([]byte(fmt.Sprintf("world %d", i))))
	}

	if err := grp.Wait(); err != nil {
		t.Error(err)
	}
}

func TestWasmBundle(t *testing.T) {
	res, err := sharedRT.Do(rt.NewJob("hello-echo", []byte("wasmWorker!"))).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "Then returned error"))
		return
	}

	if string(res.([]byte)) != "hello wasmWorker!" {
		t.Error(fmt.Errorf("expected 'hello wasmWorker!', got %s", string(res.([]byte))))
	}
}

func TestWasmLargeData(t *testing.T) {
	r := rt.New()

	doWasm := r.Register("wasm", rwasm.NewRunner("../testdata/hello-echo/hello-echo.wasm"))

	res := doWasm([]byte(largeInput))

	result, err := res.Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then for large input"))
	}

	if len(string(result.([]byte))) < 64000 {
		t.Error(fmt.Errorf("large input job . too small, got %d", len(string(result.([]byte)))))
	}

	if string(result.([]byte)) != fmt.Sprintf("hello %s", largeInput) {
		t.Error(fmt.Errorf("large input result did not match"))
	}
}

func TestWasmLargeDataGroup(t *testing.T) {
	r := rt.New()

	doWasm := r.Register("wasm", rwasm.NewRunner("../testdata/hello-echo/hello-echo.wasm"))

	grp := rt.NewGroup()
	for i := 0; i < 5000; i++ {
		grp.Add(doWasm([]byte(largeInput)))
	}

	if err := grp.Wait(); err != nil {
		t.Error("group returned an error")
	}
}

func TestWasmLargeDataGroupWithPool(t *testing.T) {
	r := rt.New()

	doWasm := r.Register("wasm", rwasm.NewRunner("../testdata/hello-echo/hello-echo.wasm"), rt.PoolSize(5))

	grp := rt.NewGroup()
	for i := 0; i < 5000; i++ {
		grp.Add(doWasm([]byte(largeInput)))
	}

	if err := grp.Wait(); err != nil {
		t.Error("group returned an error")
	}
}

func TestWasmFileGetStatic(t *testing.T) {
	getJob := rt.NewJob("get-static", "important.md")

	r, err := sharedRT.Do(getJob).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Do get-static job"))
		return
	}

	result := string(r.([]byte))

	expected := "# Hello, World\n\nContents are very important"

	if result != expected {
		t.Error("failed, got:\n", result, "\nexpeted:\n", expected)
	}
}
