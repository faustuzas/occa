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
	str := []byte(id.String())
	buff := make([]byte, len(str)+2)

	buff[0] = '"'
	copy(buff[1:len(buff)-1], str)
	buff[len(buff)-1] = '"'

	return buff, nil
}

func (id *ID) UnmarshalJSON(data []byte) error {
	u, err := uuid.ParseBytes(data)
	if err != nil {
		return err
	}

	*id = ID(u)
	return nil
}
