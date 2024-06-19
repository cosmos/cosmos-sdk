package postgres

//type graphqlHandler struct {
//	conn *pgx.Conn
//}
//
//var _ http.Handler = &graphqlHandler{}
//
//type graphqlRequest struct {
//	Query         string `json:"query"`
//	OperationName string `json:"operationName"`
//	Variables     string `json:"variables"`
//}
//
//type graphqlResponse struct {
//	Data   interface{} `json:"data"`
//	Errors []error     `json:"errors,omitempty"`
//}
//
//func (g graphqlHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
//	// decode request
//	var req graphqlRequest
//	err := json.NewDecoder(request.Body).Decode(&req)
//	if err != nil {
//		http.Error(writer, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//}
