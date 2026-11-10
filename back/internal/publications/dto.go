package publications

type ValidateResponse struct {
    Valid   bool              `json:"valid"`
    Errors  []FieldDetail     `json:"errors"`     
    Summary ValidateSummary   `json:"summary"`
}

type ValidateSummary struct {
    QuestionCount int     `json:"questionCount"`
    MaxScore      float64 `json:"maxScore"`
}

type PublishRequest struct {
    ExpectedRevision int `json:"expectedRevision"`
}

type PublishResponse struct {
    Publication PublishResponseData `json:"publication"`
}

type PublishResponseData struct {
    ID          string `json:"id"`
    Version     int    `json:"version"`
    PublicToken string `json:"publicToken"`
    PublicURL   string `json:"publicUrl"`
    PublishedAt string `json:"publishedAt"`
}

type ErrorResponse struct {
    Error ErrorBody `json:"error"`
}

type ErrorBody struct {
    Code    string        `json:"code"`
    Message string        `json:"message"`
    Details []FieldDetail `json:"details,omitempty"`   
}