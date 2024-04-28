package models

type Message struct {
	Action     string                 `json:"action,omitempty"`
	MessageId  string                 `json:"message_id,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Data       []byte                 `json:"data,omitempty"`
}

type CreateUserRequest struct {
	Name  string `json:"name,omitempty"`
	ReqId string `json:"req_id,omitempty"`
}

type CreateUserResponse struct {
	Code    uint32 `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    struct {
		ReqId string `json:"req_id,omitempty"`
		Name  string `json:"name,omitempty"`
		Id    int64  `json:"id,omitempty"`
	} `json:"data,omitempty"`
}

type UserCtx struct {
	UserId         int64 `json:"user_id,omitempty"`
	TsEnroll       int64 `json:"ts_enroll,omitempty"`
	TsInActivePool int64 `json:"ts_in_active_pool,omitempty"`
}
