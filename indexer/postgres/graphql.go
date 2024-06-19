package postgres

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type graphqlHandler struct {
	conn *sql.DB
}

func NewGraphQLHandler(conn *sql.DB) http.Handler {
	return &graphqlHandler{conn: conn}
}

var _ http.Handler = &graphqlHandler{}

type graphqlRequest struct {
	Query         string `json:"query"`
	OperationName string `json:"operationName"`
	Variables     string `json:"variables"`
}

type graphqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []error         `json:"errors,omitempty"`
}

func (g graphqlHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		g.handleGet(writer, request)
	case http.MethodPost:
		g.handlePost(writer, request)
	default:
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (g graphqlHandler) handlePost(writer http.ResponseWriter, request *http.Request) {
	var req graphqlRequest
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	g.handle(writer, request, &req)
}

func (g graphqlHandler) handleGet(writer http.ResponseWriter, request *http.Request) {
	var gqlReq graphqlRequest
	gqlReq.Query = request.URL.Query().Get("query")
	gqlReq.OperationName = request.URL.Query().Get("operationName")
	gqlReq.Variables = request.URL.Query().Get("variables")

	g.handle(writer, request, &gqlReq)
}

func (g graphqlHandler) handle(writer http.ResponseWriter, request *http.Request, gqlReq *graphqlRequest) {
	rows, err := g.conn.QueryContext(
		request.Context(),
		`select graphql.resolve(?, ?, ?)`,
		gqlReq.Query,
		gqlReq.OperationName,
		gqlReq.Variables,
	)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var data json.RawMessage
	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	resp := graphqlResponse{Data: data}
	err = json.NewEncoder(writer).Encode(resp)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
}
