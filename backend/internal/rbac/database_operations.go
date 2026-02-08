package rbac

import (
	"cloud.google.com/go/datastore"
)

// Role database operations
func (r *Role) LoadKey(k *datastore.Key) error {
	r.ID = k.Name
	if r.ID == "" && k.ID != 0 {
		r.ID = k.Encode()
	}
	return nil
}

func (r *Role) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(r)
}

func (r *Role) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(r, ps)
}

// Permission database operations
func (p *Permission) LoadKey(k *datastore.Key) error {
	p.ID = k.Name
	if p.ID == "" && k.ID != 0 {
		p.ID = k.Encode()
	}
	return nil
}

func (p *Permission) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(p)
}

func (p *Permission) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(p, ps)
}

// UserRole database operations
func (ur *UserRole) LoadKey(k *datastore.Key) error {
	ur.ID = k.Name
	if ur.ID == "" && k.ID != 0 {
		ur.ID = k.Encode()
	}
	return nil
}

func (ur *UserRole) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(ur)
}

func (ur *UserRole) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(ur, ps)
}
