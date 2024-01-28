package id

import "github.com/google/uuid"

type ID uuid.UUID

func NewID() ID {
	return ID(uuid.New())
}

func FromString(str string) ID {
	return ID(uuid.MustParse(str))
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

func (id ID) MarshalJSON() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *ID) UnmarshalJSON(data []byte) error {
	u, err := uuid.ParseBytes(data)
	if err != nil {
		return err
	}

	*id = ID(u)
	return nil
}
