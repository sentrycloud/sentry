package dbmodel

import "time"

// Entity all MySQL tables should include these columns, embed this structure to all table entity
type Entity struct {
	ID        uint64    `json:"id"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	IsDeleted int       `json:"is_deleted"`
}

func (p *Entity) SetTimeNow() {
	now := time.Now()
	p.Created = now
	p.Updated = now
}
