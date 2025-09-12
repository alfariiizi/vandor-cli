package schema

import "entgo.io/ent"

// TestSchema holds the schema definition for the TestSchema entity.
type TestSchema struct {
	ent.Schema
}

// Fields of the TestSchema.
func (TestSchema) Fields() []ent.Field {
	return nil
}

// Edges of the TestSchema.
func (TestSchema) Edges() []ent.Edge {
	return nil
}
