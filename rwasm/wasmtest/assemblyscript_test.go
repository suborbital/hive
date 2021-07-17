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

func TestASEcho(t *testing.T) {
	r := rt.New()

	// test a WASM module that is loaded directly instead of through the bundle
	doWasm := r.Register("as-echo", rwasm.NewRunner("../testdata/as-echo/as-echo.wasm"))

	res, err := doWasm("from AssemblyScript!").Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	fmt.Println(string(res.([]byte)))

	if string(res.([]byte)) != "hello, from AssemblyScript!" {
		t.Error("as-echo failed, got:", string(res.([]byte)))
	}
}

func TestASFetch(t *testing.T) {
	r := rt.New()

	// test a WASM module that is loaded directly instead of through the bundle
	doWasm := r.Register("as-fetch", rwasm.NewRunner("../testdata/as-fetch/as-fetch.wasm"))

	res, err := doWasm("https://1password.com").Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	if string(res.([]byte)[:100]) != "<!doctype html><html lang=en data-language-url=/><head><meta charset=utf-8><meta name=viewport conte" {
		t.Error("as-fetch failed, got:", string(res.([]byte)[:100]))
	}
}

func TestASJSON(t *testing.T) {
	r := rt.New()

	// test a WASM module that is loaded directly instead of through the bundle
	doWasm := r.Register("as-json", rwasm.NewRunner("../testdata/as-json/as-json.wasm"))

	res, err := doWasm("").Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	if string(res.([]byte)) != `{"firstName":"Connor","lastName":"Hicks","age":26,"meta":{"country":"Canada"},"tags":["hello","world"]}` {
		t.Error("as-json failed, got:", string(res.([]byte)))
	}
}

func TestASLargeData(t *testing.T) {
	r := rt.New()

	// test a WASM module that is loaded directly instead of through the bundle
	doWasm := r.Register("as-echo", rwasm.NewRunner("../testdata/as-echo/as-echo.wasm"))

	res, err := doWasm(largeInput).Then()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to Then"))
		return
	}

	if string(res.([]byte)) != "hello, "+largeInput {
		t.Error("as-test failed, got:", string(res.([]byte)))
	}
}

func TestASRunnerWithRequest(t *testing.T) {
	r := rt.New()

	doWasm := r.Register("wasm", rwasm.NewRunner("../testdata/as-req/as-req.wasm"))

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
