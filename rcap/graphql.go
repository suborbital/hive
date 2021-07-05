package rcap

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// inputs: {
// method: "GET" | "post"
// headers: map[string]string
// endpoint: string
// query: string,
// variables?: {},
// operationName?: string,
// }
// outputs { data: null |{}, errors: null | [{ message, path }] }

// GraphQLClient is a GraphQL capability for Reactr Modules
type GraphQLClient struct {
	client *http.Client
}

// NewGraphQLClient creates a GraphQLClient object
func NewGraphQLClient() *GraphQLClient {
	g := &GraphQLClient{
		client: http.DefaultClient,
	}

	return g
}

// GraphQLRequest is a request to a GraphQL endpoint
type GraphQLRequest struct {
	Query         string            `json:"query"`
	Variables     map[string]string `json:"variables,omitempty"`
	OperationName string            `json:"operationName,omitempty"`
}

// GraphQLResponse is a GraphQL response
type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data"`
	Errors []GraphQLError         `json:"errors,omitempty"`
}

// GraphQLError is a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
	Path    string `json:"path"`
}

func (g *GraphQLClient) Do(endpoint, query string) (*GraphQLResponse, error) {
	r := &GraphQLRequest{
		Query:     query,
		Variables: map[string]string{},
	}

	reqBytes, err := json.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Marshal request")
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewRequest")
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Do")
	}

	defer resp.Body.Close()

	respJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ReadAll body")
	}

	gqlResp := &GraphQLResponse{}
	if err := json.Unmarshal(respJSON, gqlResp); err != nil {
		return nil, errors.Wrap(err, "failed to Unmarshal response")
	}

	if resp.StatusCode > 299 {
		return gqlResp, errors.New("non-200 HTTP response code")
	}

	if gqlResp.Errors != nil && len(gqlResp.Errors) > 0 {
		return gqlResp, errors.New("GraphQL returned errors")
	}

	return gqlResp, nil
}
