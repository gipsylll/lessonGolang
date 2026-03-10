package domain

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Version int    `json:"version"`
}

type CreateUserInput struct {
	Name  string
	Email string
}

type UpdateUserInput struct {
	Name  string
	Email string
}

type PatchUserInput struct {
	Name  *string
	Email *string
}

type ListUsersInput struct {
	PageSize int
	Cursor   string // empty = first page; format: "RFC3339Nano|id"
}

type UserPage struct {
	Items      []User `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}
