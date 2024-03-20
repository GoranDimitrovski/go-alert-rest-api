package response

type NotFoundResponse struct {
    Message string `json:"message"`
}

func (r NotFoundResponse) GetStatusCode() int {
    return 404
}
